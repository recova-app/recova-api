package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/recova-app/backend-v2/internal/platform/config"
)

const (
	defaultOpenAIBaseURL = "https://api.openai.com/v1"
	defaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// Provider identifies the concrete upstream AI provider family.
type Provider string

const (
	ProviderGemini           Provider = "gemini"
	ProviderOpenAICompatible Provider = "openai-compatible"
)

// ErrorKind is a stable downstream error classification.
type ErrorKind string

const (
	ErrorKindTimeout         ErrorKind = "timeout"
	ErrorKindUnavailable     ErrorKind = "unavailable"
	ErrorKindInvalidResponse ErrorKind = "invalid_response"
	ErrorKindUnauthorized    ErrorKind = "unauthorized"
	ErrorKindUnknown         ErrorKind = "unknown"
)

// Message represents one conversational turn for provider requests.
type Message struct {
	Role    string
	Content string
}

// GenerateRequest is the normalized provider request shape for AI module use-cases.
type GenerateRequest struct {
	SystemInstruction string
	UserPrompt        string
	History           []Message
	ForceJSON         bool
	Persona           string
}

// GenerateResponse is normalized provider output.
type GenerateResponse struct {
	Provider Provider
	Model    string
	Text     string
}

// Client provides provider-agnostic text generation.
type Client interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}

// ProviderError contains safe classified downstream failure metadata.
type ProviderError struct {
	Provider   Provider
	Kind       ErrorKind
	StatusCode int
	Message    string
	Cause      error
}

// Error returns safe non-sensitive diagnostic text.
func (e *ProviderError) Error() string {
	if e == nil {
		return ""
	}
	base := strings.TrimSpace(e.Message)
	if base == "" {
		base = "provider request failed"
	}
	if strings.TrimSpace(string(e.Provider)) != "" {
		base = fmt.Sprintf("%s (%s)", base, e.Provider)
	}
	return base
}

// Unwrap returns the wrapped cause.
func (e *ProviderError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// NewClient constructs runtime AI client from validated configuration.
func NewClient(cfg config.AIConfig) (Client, error) {
	primary, err := buildProviderClient(providerRuntimeConfig{
		Provider: Provider(strings.TrimSpace(cfg.Provider)),
		Model:    strings.TrimSpace(cfg.Model),
		APIKey:   strings.TrimSpace(cfg.APIKey),
		BaseURL:  strings.TrimSpace(cfg.BaseURL),
		Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}

	client := &compositeClient{primary: primary}

	fallbackProvider := strings.TrimSpace(cfg.FallbackProvider)
	fallbackModel := strings.TrimSpace(cfg.FallbackModel)
	if fallbackProvider == "" || fallbackModel == "" {
		return client, nil
	}

	fallback, err := buildProviderClient(providerRuntimeConfig{
		Provider: Provider(fallbackProvider),
		Model:    fallbackModel,
		APIKey:   strings.TrimSpace(cfg.APIKey),
		BaseURL:  strings.TrimSpace(cfg.BaseURL),
		Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
	})
	if err != nil {
		return nil, err
	}

	client.fallback = fallback
	return client, nil
}

// ClassifyError returns best-effort stable downstream error kind.
func ClassifyError(err error) ErrorKind {
	if err == nil {
		return ErrorKindUnknown
	}
	var pErr *ProviderError
	if errors.As(err, &pErr) {
		return pErr.Kind
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrorKindTimeout
	}
	if isNetworkTimeout(err) {
		return ErrorKindTimeout
	}
	return ErrorKindUnknown
}

// IsFallbackEligible returns true when fallback provider should be attempted.
func IsFallbackEligible(err error) bool {
	kind := ClassifyError(err)
	return kind == ErrorKindTimeout || kind == ErrorKindUnavailable
}

type providerClient interface {
	Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error)
}

type providerRuntimeConfig struct {
	Provider Provider
	Model    string
	APIKey   string
	BaseURL  string
	Timeout  time.Duration
}

type compositeClient struct {
	primary  providerClient
	fallback providerClient
}

func (c *compositeClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if c == nil || c.primary == nil {
		return GenerateResponse{}, &ProviderError{Kind: ErrorKindUnavailable, Message: "provider client is not ready"}
	}

	resp, err := c.primary.Generate(ctx, req)
	if err == nil || c.fallback == nil || !IsFallbackEligible(err) {
		return resp, err
	}

	fallbackResp, fallbackErr := c.fallback.Generate(ctx, req)
	if fallbackErr == nil {
		return fallbackResp, nil
	}
	return GenerateResponse{}, fallbackErr
}

func buildProviderClient(cfg providerRuntimeConfig) (providerClient, error) {
	provider := Provider(strings.TrimSpace(string(cfg.Provider)))
	model := strings.TrimSpace(cfg.Model)
	apiKey := strings.TrimSpace(cfg.APIKey)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	if model == "" {
		return nil, fmt.Errorf("ai model is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("ai api key is required")
	}

	httpClient := &http.Client{}

	switch provider {
	case ProviderOpenAICompatible:
		base := cfg.BaseURL
		if strings.TrimSpace(base) == "" {
			base = defaultOpenAIBaseURL
		}
		parsed, err := normalizeBaseURL(base)
		if err != nil {
			return nil, fmt.Errorf("invalid openai-compatible base url: %w", err)
		}
		return &openAICompatibleClient{
			httpClient: httpClient,
			provider:   provider,
			model:      model,
			apiKey:     apiKey,
			baseURL:    parsed,
			timeout:    timeout,
		}, nil
	case ProviderGemini:
		base := cfg.BaseURL
		if strings.TrimSpace(base) == "" {
			base = defaultGeminiBaseURL
		}
		parsed, err := normalizeBaseURL(base)
		if err != nil {
			return nil, fmt.Errorf("invalid gemini base url: %w", err)
		}
		return &geminiClient{
			httpClient: httpClient,
			provider:   provider,
			model:      model,
			apiKey:     apiKey,
			baseURL:    parsed,
			timeout:    timeout,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported ai provider: %s", provider)
	}
}

type openAICompatibleClient struct {
	httpClient *http.Client
	provider   Provider
	model      string
	apiKey     string
	baseURL    string
	timeout    time.Duration
}

func (c *openAICompatibleClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	messages := make([]openAIMessage, 0, len(req.History)+2)
	if system := strings.TrimSpace(req.SystemInstruction); system != "" {
		messages = append(messages, openAIMessage{Role: "system", Content: system})
	}
	for _, item := range req.History {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		messages = append(messages, openAIMessage{Role: mapOpenAIRole(item.Role), Content: content})
	}
	messages = append(messages, openAIMessage{Role: "user", Content: strings.TrimSpace(req.UserPrompt)})

	payload := openAIRequest{
		Model:    c.model,
		Messages: messages,
	}
	if req.ForceJSON {
		payload.ResponseFormat = &openAIResponseFormat{Type: "json_object"}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, Message: "failed to encode provider request", Cause: err}
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, Message: "failed to build provider request", Cause: err}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, classifyTransportError(c.provider, err)
	}
	defer httpResp.Body.Close()

	respBody, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "failed to read provider response", Cause: readErr}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return GenerateResponse{}, classifyHTTPError(c.provider, httpResp.StatusCode, extractOpenAIErrorMessage(respBody))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "invalid provider response payload", Cause: err}
	}

	if len(parsed.Choices) == 0 {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "provider response has no choices"}
	}

	text := extractOpenAIContent(parsed.Choices[0].Message.Content)
	if strings.TrimSpace(text) == "" {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "provider response content is empty"}
	}

	return GenerateResponse{Provider: c.provider, Model: c.model, Text: text}, nil
}

type geminiClient struct {
	httpClient *http.Client
	provider   Provider
	model      string
	apiKey     string
	baseURL    string
	timeout    time.Duration
}

func (c *geminiClient) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	contents := make([]geminiContent, 0, len(req.History)+1)
	for _, item := range req.History {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		contents = append(contents, geminiContent{
			Role:  mapGeminiRole(item.Role),
			Parts: []geminiPart{{Text: content}},
		})
	}
	contents = append(contents, geminiContent{Role: "user", Parts: []geminiPart{{Text: strings.TrimSpace(req.UserPrompt)}}})

	payload := geminiRequest{Contents: contents}
	if system := strings.TrimSpace(req.SystemInstruction); system != "" {
		payload.SystemInstruction = &geminiSystemInstruction{Parts: []geminiPart{{Text: system}}}
	}
	if req.ForceJSON {
		payload.GenerationConfig = &geminiGenerationConfig{ResponseMimeType: "application/json"}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, Message: "failed to encode provider request", Cause: err}
	}

	endpoint := strings.TrimRight(c.baseURL, "/") + "/models/" + c.model + ":generateContent"
	httpReq, err := http.NewRequestWithContext(requestCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, Message: "failed to build provider request", Cause: err}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return GenerateResponse{}, classifyTransportError(c.provider, err)
	}
	defer httpResp.Body.Close()

	respBody, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "failed to read provider response", Cause: readErr}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return GenerateResponse{}, classifyHTTPError(c.provider, httpResp.StatusCode, extractGeminiErrorMessage(respBody))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "invalid provider response payload", Cause: err}
	}
	if len(parsed.Candidates) == 0 {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "provider response has no candidates"}
	}

	text := extractGeminiText(parsed.Candidates[0])
	if strings.TrimSpace(text) == "" {
		return GenerateResponse{}, &ProviderError{Provider: c.provider, Kind: ErrorKindInvalidResponse, StatusCode: httpResp.StatusCode, Message: "provider response content is empty"}
	}

	return GenerateResponse{Provider: c.provider, Model: c.model, Text: text}, nil
}

func normalizeBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("absolute url is required")
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func classifyTransportError(provider Provider, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || isNetworkTimeout(err) {
		return &ProviderError{Provider: provider, Kind: ErrorKindTimeout, Message: "provider request timed out", Cause: err}
	}
	if errors.Is(err, context.Canceled) {
		return &ProviderError{Provider: provider, Kind: ErrorKindUnavailable, Message: "provider request canceled", Cause: err}
	}
	return &ProviderError{Provider: provider, Kind: ErrorKindUnavailable, Message: "provider request failed", Cause: err}
}

func classifyHTTPError(provider Provider, statusCode int, providerMessage string) error {
	message := strings.TrimSpace(providerMessage)
	if message == "" {
		message = "provider request failed"
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &ProviderError{Provider: provider, Kind: ErrorKindUnauthorized, StatusCode: statusCode, Message: message}
	case http.StatusBadRequest, http.StatusNotFound, http.StatusUnprocessableEntity:
		return &ProviderError{Provider: provider, Kind: ErrorKindInvalidResponse, StatusCode: statusCode, Message: message}
	case http.StatusTooManyRequests, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return &ProviderError{Provider: provider, Kind: ErrorKindUnavailable, StatusCode: statusCode, Message: message}
	case http.StatusInternalServerError, http.StatusBadGateway:
		return &ProviderError{Provider: provider, Kind: ErrorKindUnavailable, StatusCode: statusCode, Message: message}
	default:
		if statusCode >= 500 {
			return &ProviderError{Provider: provider, Kind: ErrorKindUnavailable, StatusCode: statusCode, Message: message}
		}
		return &ProviderError{Provider: provider, Kind: ErrorKindInvalidResponse, StatusCode: statusCode, Message: message}
	}
}

func isNetworkTimeout(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return true
	}
	return false
}

func mapOpenAIRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant", "model":
		return "assistant"
	case "system", "developer":
		return "system"
	default:
		return "user"
	}
}

func mapGeminiRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "assistant", "model":
		return "model"
	default:
		return "user"
	}
}

func extractOpenAIContent(raw json.RawMessage) string {
	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return strings.TrimSpace(plain)
	}

	var parts []openAIContentPart
	if err := json.Unmarshal(raw, &parts); err == nil {
		chunks := make([]string, 0, len(parts))
		for _, part := range parts {
			if strings.EqualFold(strings.TrimSpace(part.Type), "text") {
				text := strings.TrimSpace(part.Text)
				if text != "" {
					chunks = append(chunks, text)
				}
			}
		}
		return strings.Join(chunks, "\n")
	}

	return ""
}

func extractGeminiText(candidate geminiCandidate) string {
	chunks := make([]string, 0, len(candidate.Content.Parts))
	for _, part := range candidate.Content.Parts {
		text := strings.TrimSpace(part.Text)
		if text == "" {
			continue
		}
		chunks = append(chunks, text)
	}
	return strings.Join(chunks, "\n")
}

func extractOpenAIErrorMessage(body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
			return msg
		}
	}
	return ""
}

func extractGeminiErrorMessage(body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
			return msg
		}
	}
	return ""
}

type openAIRequest struct {
	Model          string                `json:"model"`
	Messages       []openAIMessage       `json:"messages"`
	ResponseFormat *openAIResponseFormat `json:"response_format,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Message openAIChoiceMessage `json:"message"`
}

type openAIChoiceMessage struct {
	Content json.RawMessage `json:"content"`
}

type openAIContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  *geminiGenerationConfig  `json:"generationConfig,omitempty"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}
