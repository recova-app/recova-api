package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	e2eharness "github.com/recova-app/backend-v2/test/harness/e2e"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

type criticalFlowStep struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type criticalFlowReport struct {
	Suite       string             `json:"suite"`
	Scope       string             `json:"scope"`
	GeneratedAt string             `json:"generated_at"`
	DurationMs  int64              `json:"duration_ms"`
	Status      string             `json:"status"`
	Steps       []criticalFlowStep `json:"steps"`
}

func TestE2E_CriticalFlows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E suite in short mode")
	}

	runtime := e2eharness.NewRuntime(t)
	steps := make([]criticalFlowStep, 0, 16)
	start := time.Now()
	reportStatus := "passed"
	scope := normalizeE2EScope(os.Getenv("RECOVA_E2E_SCOPE"))

	t.Cleanup(func() {
		report := criticalFlowReport{
			Suite:       "critical-flows",
			Scope:       scope,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			DurationMs:  time.Since(start).Milliseconds(),
			Status:      reportStatus,
			Steps:       steps,
		}
		writeJSONReport(t, os.Getenv("RECOVA_E2E_REPORT_PATH"), report)
	})

	var access_token string
	var refreshCookie string
	var communityPostID string
	var communityCommentID string

	runStep := func(name string, fn func() error) {
		stepIndex := len(steps)
		steps = append(steps, criticalFlowStep{Name: name, Status: "running"})
		if !e2eScopeAllowsStep(scope, name) {
			steps[stepIndex].Status = "skipped"
			steps[stepIndex].Detail = "excluded by RECOVA_E2E_SCOPE"
			return
		}
		if err := fn(); err != nil {
			reportStatus = "failed"
			steps[stepIndex].Status = "failed"
			steps[stepIndex].Detail = err.Error()
			t.Fatalf("%s: %v", name, err)
		}
		steps[stepIndex].Status = "passed"
	}

	runStep("health readiness", func() error {
		resp := sendJSONRequest(t, runtime, http.MethodGet, "/health/ready", nil, nil)
		if resp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("expected 200 got %d body=%s", resp.StatusCode, string(resp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, resp.JSON)
		return nil
	})

	runStep("auth login flow", func() error {
		resp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/google", map[string]any{
			"token": e2eharness.ValidGoogleToken(),
		}, nil)
		if resp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("expected 200 got %d body=%s", resp.StatusCode, string(resp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, resp.JSON)

		access_token = readSessionAccessToken(resp.JSON)
		if strings.TrimSpace(access_token) == "" {
			return fmt.Errorf("access token is empty")
		}

		refreshCookie = extractCookiePair(resp.Header, "recova_refresh_e2e")
		if strings.TrimSpace(refreshCookie) == "" {
			return fmt.Errorf("refresh cookie not found")
		}
		return nil
	})

	runStep("auth refresh token flow", func() error {
		resp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/refresh", nil, map[string]string{
			"Cookie": refreshCookie,
		})
		if resp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("expected 200 got %d body=%s", resp.StatusCode, string(resp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, resp.JSON)

		access_token = readSessionAccessToken(resp.JSON)
		if strings.TrimSpace(access_token) == "" {
			return fmt.Errorf("refreshed access token is empty")
		}

		newCookie := extractCookiePair(resp.Header, "recova_refresh_e2e")
		if strings.TrimSpace(newCookie) != "" {
			refreshCookie = newCookie
		}
		return nil
	})

	runStep("auth manual register/login flow", func() error {
		email := "e2e-manual@example.test"
		username := "e2e_manual_user"
		password := "password123"

		registerResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/register", map[string]any{
			"email":            email,
			"username":         username,
			"password":         password,
			"confirm_password": password,
		}, nil)
		if registerResp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("manual register expected 201 got %d body=%s", registerResp.StatusCode, string(registerResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, registerResp.JSON)

		loginResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"identifier": email,
			"password":   password,
		}, nil)
		if loginResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("manual login expected 200 got %d body=%s", loginResp.StatusCode, string(loginResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, loginResp.JSON)

		manualAccessToken := readSessionAccessToken(loginResp.JSON)
		if strings.TrimSpace(manualAccessToken) == "" {
			return fmt.Errorf("manual access token is empty")
		}
		manualRefreshCookie := extractCookiePair(loginResp.Header, "recova_refresh_e2e")
		if strings.TrimSpace(manualRefreshCookie) == "" {
			return fmt.Errorf("manual refresh cookie not found")
		}

		meResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/users/me", nil, bearerHeaders(manualAccessToken))
		if meResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("users me expected 200 got %d body=%s", meResp.StatusCode, string(meResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, meResp.JSON)

		logoutResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/logout", nil, mergeHeaders(bearerHeaders(manualAccessToken), map[string]string{
			"Cookie": manualRefreshCookie,
		}))
		if logoutResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("manual logout expected 200 got %d body=%s", logoutResp.StatusCode, string(logoutResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, logoutResp.JSON)

		refreshResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/refresh", nil, map[string]string{
			"Cookie": manualRefreshCookie,
		})
		if refreshResp.StatusCode != fiber.StatusUnauthorized {
			return fmt.Errorf("manual refresh after logout expected 401 got %d body=%s", refreshResp.StatusCode, string(refreshResp.Body))
		}
		httpharness.RequireErrorEnvelope(t, refreshResp.JSON, "UNAUTHENTICATED")
		return nil
	})

	runStep("onboarding and profile", func() error {
		resp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/onboarding", map[string]any{
			"nickname":           "e2e-user",
			"recovery_reason":    "ingin pulih konsisten",
			"daily_checkin_time": "08:30",
			"porn_free_goal":     7,
			"answers": map[string]any{
				"motivation": "keluarga",
			},
			"dependency_level": "Sedang",
		}, bearerHeaders(access_token))
		if resp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("expected 201 got %d body=%s", resp.StatusCode, string(resp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, resp.JSON)

		meResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/users/me", nil, bearerHeaders(access_token))
		if meResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("users/me expected 200 got %d body=%s", meResp.StatusCode, string(meResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, meResp.JSON)
		if !readOnboardingCompleted(meResp.JSON) {
			return fmt.Errorf("expected onboarding_completed=true")
		}
		return nil
	})

	runStep("daily checkin and statistics", func() error {
		checkinResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/routine/checkin", map[string]any{
			"mood":          "tenang",
			"is_successful": true,
			"commitment":    "fokus 10 menit",
		}, bearerHeaders(access_token))
		if checkinResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("checkin expected 200 got %d body=%s", checkinResp.StatusCode, string(checkinResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, checkinResp.JSON)

		statsResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/routine/statistics", nil, bearerHeaders(access_token))
		if statsResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("statistics expected 200 got %d body=%s", statsResp.StatusCode, string(statsResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, statsResp.JSON)

		total_checkins := readNestedFloat(statsResp.JSON, "data", "total_checkins")
		if total_checkins < 1 {
			return fmt.Errorf("expected total_checkins >= 1, got %.0f", total_checkins)
		}

		summaryResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/routine/statistics/activity-summary?window_days=30", nil, bearerHeaders(access_token))
		if summaryResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("activity summary expected 200 got %d body=%s", summaryResp.StatusCode, string(summaryResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, summaryResp.JSON)
		if readNestedFloat(summaryResp.JSON, "data", "window_days") != 30 {
			return fmt.Errorf("activity summary window_days mismatch")
		}
		return nil
	})

	runStep("achievements catalog/progress/unlocked", func() error {
		catalogResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/achievements/catalog", nil, bearerHeaders(access_token))
		if catalogResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("achievements catalog expected 200 got %d body=%s", catalogResp.StatusCode, string(catalogResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, catalogResp.JSON)

		progressResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/achievements/progress", nil, bearerHeaders(access_token))
		if progressResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("achievements progress expected 200 got %d body=%s", progressResp.StatusCode, string(progressResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, progressResp.JSON)

		unlockedResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/achievements/unlocked", nil, bearerHeaders(access_token))
		if unlockedResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("achievements unlocked expected 200 got %d body=%s", unlockedResp.StatusCode, string(unlockedResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, unlockedResp.JSON)
		return nil
	})

	runStep("journals create/list", func() error {
		createResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/journals", map[string]any{
			"content": "today I successfully resisted the urge",
		}, bearerHeaders(access_token))
		if createResp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("journals create expected 201 got %d body=%s", createResp.StatusCode, string(createResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, createResp.JSON)

		listResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/journals", nil, bearerHeaders(access_token))
		if listResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("journals list expected 200 got %d body=%s", listResp.StatusCode, string(listResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, listResp.JSON)
		if len(readDataArray(listResp.JSON)) < 1 {
			return fmt.Errorf("expected journal list not empty")
		}
		return nil
	})

	runStep("community post/comment/like", func() error {
		createResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/community", map[string]any{
			"content":  "stay strong, one day at a time",
			"category": "motivasi",
		}, bearerHeaders(access_token))
		if createResp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("community create expected 201 got %d body=%s", createResp.StatusCode, string(createResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, createResp.JSON)
		communityPostID = readNestedString(createResp.JSON, "data", "id")
		if strings.TrimSpace(communityPostID) == "" {
			return fmt.Errorf("post id is empty")
		}

		commentResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/community/"+communityPostID+"/comments", map[string]any{
			"content": "supportive comment",
		}, bearerHeaders(access_token))
		if commentResp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("community comment expected 201 got %d body=%s", commentResp.StatusCode, string(commentResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, commentResp.JSON)
		communityCommentID = readNestedString(commentResp.JSON, "data", "id")
		if strings.TrimSpace(communityCommentID) == "" {
			return fmt.Errorf("comment id is empty")
		}

		replyResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/community/"+communityPostID+"/comments/"+communityCommentID+"/replies", map[string]any{
			"content": "reply dukungan lanjutan",
		}, bearerHeaders(access_token))
		if replyResp.StatusCode != fiber.StatusCreated {
			return fmt.Errorf("community reply expected 201 got %d body=%s", replyResp.StatusCode, string(replyResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, replyResp.JSON)
		if readNestedString(replyResp.JSON, "data", "parent_comment_id") != communityCommentID {
			return fmt.Errorf("reply parent_comment_id mismatch")
		}

		threadResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/community/"+communityPostID+"/comments?limit=50", nil, bearerHeaders(access_token))
		if threadResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("community thread expected 200 got %d body=%s", threadResp.StatusCode, string(threadResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, threadResp.JSON)
		comments := readNestedArray(threadResp.JSON, "data", "comments")
		if len(comments) < 1 {
			return fmt.Errorf("community thread comments are empty")
		}
		root, ok := comments[0].(map[string]any)
		if !ok {
			return fmt.Errorf("community thread root is invalid")
		}
		reply_count, _ := root["reply_count"].(float64)
		if int(reply_count) < 1 {
			return fmt.Errorf("community thread reply_count did not increase")
		}

		likeResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/community/"+communityPostID+"/like", nil, bearerHeaders(access_token))
		if likeResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("community like expected 200 got %d body=%s", likeResp.StatusCode, string(likeResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, likeResp.JSON)
		return nil
	})

	runStep("education and daily content read", func() error {
		educationResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/education", nil, bearerHeaders(access_token))
		if educationResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("education list expected 200 got %d body=%s", educationResp.StatusCode, string(educationResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, educationResp.JSON)
		if len(readDataArray(educationResp.JSON)) < 1 {
			return fmt.Errorf("expected education list not empty")
		}

		contentResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/content/daily", nil, bearerHeaders(access_token))
		if contentResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("daily content expected 200 got %d body=%s", contentResp.StatusCode, string(contentResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, contentResp.JSON)
		if strings.TrimSpace(readNestedString(contentResp.JSON, "data", "motivation")) == "" {
			return fmt.Errorf("motivation is empty")
		}
		if strings.TrimSpace(readNestedString(contentResp.JSON, "data", "challenge", "title")) == "" {
			return fmt.Errorf("challenge title is empty")
		}
		if strings.TrimSpace(readNestedString(contentResp.JSON, "data", "challenge", "description")) == "" {
			return fmt.Errorf("challenge description is empty")
		}
		return nil
	})

	runStep("ai coach safe response and history", func() error {
		getPersonaResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/ai/persona-preferences", nil, bearerHeaders(access_token))
		if getPersonaResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("persona preference get expected 200 got %d body=%s", getPersonaResp.StatusCode, string(getPersonaResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, getPersonaResp.JSON)

		updatePersonaResp := sendJSONRequest(t, runtime, http.MethodPut, "/api/v1/ai/persona-preferences", map[string]any{
			"persona": "direct",
		}, bearerHeaders(access_token))
		if updatePersonaResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("persona preference put expected 200 got %d body=%s", updatePersonaResp.StatusCode, string(updatePersonaResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, updatePersonaResp.JSON)
		if readNestedString(updatePersonaResp.JSON, "data", "persona") != "direct" {
			return fmt.Errorf("persona update mismatch")
		}

		askResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/ai/ask-coach", map[string]any{
			"message": "I feel close to relapse",
		}, bearerHeaders(access_token))
		if askResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("ask coach expected 200 got %d body=%s", askResp.StatusCode, string(askResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, askResp.JSON)

		coachText := strings.ToLower(strings.TrimSpace(readNestedString(askResp.JSON, "data", "response")))
		if coachText == "" {
			return fmt.Errorf("AI response is empty")
		}
		if strings.Contains(coachText, "token-e2e") {
			return fmt.Errorf("response ai memuat token sensitif")
		}
		if readNestedString(askResp.JSON, "data", "persona_used") != "direct" {
			return fmt.Errorf("persona_used mismatch")
		}

		historyResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/ai/chat-history", nil, bearerHeaders(access_token))
		if historyResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("chat history expected 200 got %d body=%s", historyResp.StatusCode, string(historyResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, historyResp.JSON)
		if len(readDataArray(historyResp.JSON)) < 2 {
			return fmt.Errorf("expected chat history has >= 2 records")
		}

		summaryResp := sendJSONRequest(t, runtime, http.MethodGet, "/api/v1/ai/summary", nil, bearerHeaders(access_token))
		if summaryResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("summary expected 200 got %d body=%s", summaryResp.StatusCode, string(summaryResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, summaryResp.JSON)
		if strings.TrimSpace(readNestedString(summaryResp.JSON, "data", "summary")) == "" {
			return fmt.Errorf("summary is empty")
		}

		analysisResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/ai/onboarding-analysis", map[string]any{
			"answers": map[string]any{"q1": "hard to sleep when alone"},
		}, bearerHeaders(access_token))
		if analysisResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("onboarding analysis expected 200 got %d body=%s", analysisResp.StatusCode, string(analysisResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, analysisResp.JSON)
		for _, key := range []string{"level", "title", "level_description", "pattern_analysis", "encouragement"} {
			if strings.TrimSpace(readNestedString(analysisResp.JSON, "data", key)) == "" {
				return fmt.Errorf("analysis field is empty: %s", key)
			}
		}

		return nil
	})

	runStep("auth logout and refresh invalidation", func() error {
		logoutResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/logout", nil, mergeHeaders(bearerHeaders(access_token), map[string]string{"Cookie": refreshCookie}))
		if logoutResp.StatusCode != fiber.StatusOK {
			return fmt.Errorf("logout expected 200 got %d body=%s", logoutResp.StatusCode, string(logoutResp.Body))
		}
		httpharness.RequireSuccessEnvelope(t, logoutResp.JSON)

		refreshResp := sendJSONRequest(t, runtime, http.MethodPost, "/api/v1/auth/refresh", nil, map[string]string{"Cookie": refreshCookie})
		if refreshResp.StatusCode != fiber.StatusUnauthorized {
			return fmt.Errorf("refresh after logout expected 401 got %d body=%s", refreshResp.StatusCode, string(refreshResp.Body))
		}
		httpharness.RequireErrorEnvelope(t, refreshResp.JSON, "UNAUTHENTICATED")
		return nil
	})
}

type requestResult struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	JSON       map[string]any
}

func sendJSONRequest(t testing.TB, runtime *e2eharness.Runtime, method string, path string, body any, headers map[string]string) requestResult {
	t.Helper()

	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := runtime.Server.Test(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	result := requestResult{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: bodyBytes}
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "application/json") {
		var payload map[string]any
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			t.Fatalf("parse json response: %v body=%s", err, string(bodyBytes))
		}
		result.JSON = payload
	}
	return result
}

func bearerHeaders(access_token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + strings.TrimSpace(access_token),
	}
}

func mergeHeaders(left map[string]string, right map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range left {
		result[key] = value
	}
	for key, value := range right {
		result[key] = value
	}
	return result
}

func readSessionAccessToken(payload map[string]any) string {
	data, ok := payload["data"].(map[string]any)
	if !ok {
		return ""
	}
	session, ok := data["session"].(map[string]any)
	if !ok {
		return ""
	}
	value, _ := session["access_token"].(string)
	return strings.TrimSpace(value)
}

func readOnboardingCompleted(payload map[string]any) bool {
	data, ok := payload["data"].(map[string]any)
	if !ok {
		return false
	}
	value, _ := data["onboarding_completed"].(bool)
	return value
}

func readNestedString(payload map[string]any, keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	var current any = payload
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = asMap[key]
		if !ok {
			return ""
		}
	}
	value, _ := current.(string)
	return strings.TrimSpace(value)
}

func readNestedFloat(payload map[string]any, keys ...string) float64 {
	if len(keys) == 0 {
		return 0
	}
	var current any = payload
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return 0
		}
		current, ok = asMap[key]
		if !ok {
			return 0
		}
	}
	value, _ := current.(float64)
	return value
}

func readDataArray(payload map[string]any) []any {
	data, ok := payload["data"].([]any)
	if !ok {
		return []any{}
	}
	return data
}

func readNestedArray(payload map[string]any, keys ...string) []any {
	if len(keys) == 0 {
		return []any{}
	}
	var current any = payload
	for _, key := range keys {
		asMap, ok := current.(map[string]any)
		if !ok {
			return []any{}
		}
		current, ok = asMap[key]
		if !ok {
			return []any{}
		}
	}
	value, ok := current.([]any)
	if !ok {
		return []any{}
	}
	return value
}

func extractCookiePair(header http.Header, cookieName string) string {
	for _, raw := range header.Values("Set-Cookie") {
		parts := strings.Split(raw, ";")
		if len(parts) == 0 {
			continue
		}
		pair := strings.TrimSpace(parts[0])
		if strings.HasPrefix(pair, strings.TrimSpace(cookieName)+"=") {
			return pair
		}
	}
	return ""
}

func writeJSONReport(t testing.TB, path string, payload any) {
	t.Helper()

	targetPath := strings.TrimSpace(path)
	if targetPath == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("create report directory: %v", err)
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}
	encoded = append(encoded, '\n')

	if err := os.WriteFile(targetPath, encoded, 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
}

func normalizeE2EScope(raw string) string {
	scope := strings.ToLower(strings.TrimSpace(raw))
	switch scope {
	case "", "all", "wave64", "wave65", "wave66", "wave67", "wave68":
		if scope == "" {
			return "all"
		}
		return scope
	default:
		return "all"
	}
}

func e2eScopeAllowsStep(scope string, stepName string) bool {
	allowedSteps := map[string]map[string]struct{}{
		"all": {},
		"wave64": {
			"health readiness": {},
		},
		"wave65": {
			"health readiness":                     {},
			"auth login flow":                      {},
			"auth refresh token flow":              {},
			"auth manual register/login flow":      {},
			"onboarding and profile":               {},
			"auth logout and refresh invalidation": {},
		},
		"wave66": {
			"health readiness":                       {},
			"auth login flow":                        {},
			"onboarding and profile":                 {},
			"daily checkin and statistics":           {},
			"achievements catalog/progress/unlocked": {},
			"journals create/list":                   {},
			"auth logout and refresh invalidation":   {},
		},
		"wave67": {
			"health readiness":                     {},
			"auth login flow":                      {},
			"onboarding and profile":               {},
			"community post/comment/like":          {},
			"education and daily content read":     {},
			"auth logout and refresh invalidation": {},
		},
		"wave68": {
			"health readiness":                     {},
			"auth login flow":                      {},
			"onboarding and profile":               {},
			"ai coach safe response and history":   {},
			"auth logout and refresh invalidation": {},
		},
	}

	selected, exists := allowedSteps[scope]
	if !exists {
		return true
	}
	if scope == "all" {
		return true
	}
	_, ok := selected[stepName]
	return ok
}
