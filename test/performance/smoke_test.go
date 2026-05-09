package performance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	e2eharness "github.com/recova-app/backend-v2/test/harness/e2e"
	httpharness "github.com/recova-app/backend-v2/test/harness/http"
)

const (
	loadRequestsPerScenario = 160
	loadWorkers             = 8
	maxErrorRateThreshold   = 0.01
	maxP95MsThreshold       = 300.0
)

type perfScenarioResult struct {
	Name      string  `json:"name"`
	Requests  int     `json:"requests"`
	Failures  int     `json:"failures"`
	ErrorRate float64 `json:"errorRate"`
	P50Ms     float64 `json:"p50Ms"`
	P95Ms     float64 `json:"p95Ms"`
	MaxMs     float64 `json:"maxMs"`
}

type perfSmokeReport struct {
	Suite           string               `json:"suite"`
	GeneratedAt     string               `json:"generatedAt"`
	DurationMs      int64                `json:"durationMs"`
	Status          string               `json:"status"`
	Thresholds      map[string]float64   `json:"thresholds"`
	ReadinessChecks map[string]int64     `json:"readinessChecks"`
	Scenarios       []perfScenarioResult `json:"scenarios"`
}

type perfScenario struct {
	name           string
	method         string
	path           string
	body           any
	expectedStatus int
	headers        map[string]string
}

func TestPerformance_LoadSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance smoke in short mode")
	}

	runtime := e2eharness.NewRuntime(t)
	start := time.Now()
	reportStatus := "passed"

	accessToken, _, err := loginAndRefresh(t, runtime)
	if err != nil {
		t.Fatalf("prepare auth session for performance test: %v", err)
	}
	communityPostID, err := prepareCommunityThread(runtime, accessToken)
	if err != nil {
		t.Fatalf("prepare community thread for performance test: %v", err)
	}

	scenarios := []perfScenario{
		{name: "health_live", method: http.MethodGet, path: "/health/live", expectedStatus: fiber.StatusOK},
		{name: "health_ready", method: http.MethodGet, path: "/health/ready", expectedStatus: fiber.StatusOK},
		{name: "users_me", method: http.MethodGet, path: "/api/v1/users/me", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "daily_content", method: http.MethodGet, path: "/api/v1/content/daily", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "routine_activity_summary", method: http.MethodGet, path: "/api/v1/routine/statistics/activity-summary?windowDays=30", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "achievements_catalog", method: http.MethodGet, path: "/api/v1/achievements/catalog", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "community_comment_thread", method: http.MethodGet, path: "/api/v1/community/" + communityPostID + "/comments?limit=50", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "ai_persona_preferences", method: http.MethodGet, path: "/api/v1/ai/persona-preferences", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
		{name: "ai_summary", method: http.MethodGet, path: "/api/v1/ai/summary", expectedStatus: fiber.StatusOK, headers: bearerHeaders(accessToken)},
	}

	probeCtx, cancelProbe := context.WithCancel(context.Background())
	var readinessSuccess int64
	var readinessFailure int64
	go readinessProbeLoop(probeCtx, runtime, &readinessSuccess, &readinessFailure)

	results := make([]perfScenarioResult, 0, len(scenarios))
	for _, scenario := range scenarios {
		result := runPerfScenario(t, runtime, scenario, loadRequestsPerScenario, loadWorkers)
		results = append(results, result)

		if result.ErrorRate > maxErrorRateThreshold {
			reportStatus = "failed"
			t.Fatalf("scenario %s errorRate %.4f exceeds threshold %.4f", scenario.name, result.ErrorRate, maxErrorRateThreshold)
		}
		if result.P95Ms > maxP95MsThreshold {
			reportStatus = "failed"
			t.Fatalf("scenario %s p95 %.2fms exceeds threshold %.2fms", scenario.name, result.P95Ms, maxP95MsThreshold)
		}
	}

	cancelProbe()
	if atomic.LoadInt64(&readinessFailure) > 0 {
		reportStatus = "failed"
		t.Fatalf("readiness probe failures detected during load: %d", atomic.LoadInt64(&readinessFailure))
	}

	report := perfSmokeReport{
		Suite:       "load-smoke",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		DurationMs:  time.Since(start).Milliseconds(),
		Status:      reportStatus,
		Thresholds: map[string]float64{
			"maxErrorRate": maxErrorRateThreshold,
			"maxP95Ms":     maxP95MsThreshold,
		},
		ReadinessChecks: map[string]int64{
			"success": atomic.LoadInt64(&readinessSuccess),
			"failure": atomic.LoadInt64(&readinessFailure),
		},
		Scenarios: results,
	}
	writeJSONReport(t, os.Getenv("RECOVA_PERF_REPORT_PATH"), report)
}

func runPerfScenario(t testing.TB, runtime *e2eharness.Runtime, scenario perfScenario, totalRequests int, workers int) perfScenarioResult {
	t.Helper()

	if workers <= 0 {
		workers = 1
	}
	if totalRequests <= 0 {
		totalRequests = 1
	}

	jobs := make(chan struct{}, totalRequests)
	for i := 0; i < totalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var mu sync.Mutex
	durations := make([]float64, 0, totalRequests)
	failures := 0

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				started := time.Now()
				resp := sendRequest(t, runtime, scenario.method, scenario.path, scenario.body, scenario.headers)
				durationMs := float64(time.Since(started).Microseconds()) / 1000.0

				mu.Lock()
				durations = append(durations, durationMs)
				if resp.StatusCode != scenario.expectedStatus {
					failures++
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	sort.Float64s(durations)
	result := perfScenarioResult{
		Name:      scenario.name,
		Requests:  len(durations),
		Failures:  failures,
		ErrorRate: float64(failures) / float64(len(durations)),
		P50Ms:     percentile(durations, 50),
		P95Ms:     percentile(durations, 95),
		MaxMs:     maxDuration(durations),
	}
	return result
}

func readinessProbeLoop(ctx context.Context, runtime *e2eharness.Runtime, success *int64, failure *int64) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp := sendRequest(nil, runtime, http.MethodGet, "/health/ready", nil, nil)
			if resp.StatusCode == fiber.StatusOK {
				atomic.AddInt64(success, 1)
				continue
			}
			atomic.AddInt64(failure, 1)
		}
	}
}

type httpResult struct {
	StatusCode int
	Header     http.Header
	JSON       map[string]any
	Body       []byte
}

func sendRequest(t testing.TB, runtime *e2eharness.Runtime, method string, path string, body any, headers map[string]string) httpResult {
	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			if t != nil {
				t.Fatalf("marshal request body: %v", err)
			}
			return httpResult{StatusCode: fiber.StatusInternalServerError}
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
		if t != nil {
			t.Fatalf("execute request: %v", err)
		}
		return httpResult{StatusCode: fiber.StatusInternalServerError}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if t != nil {
			t.Fatalf("read response body: %v", err)
		}
		return httpResult{StatusCode: fiber.StatusInternalServerError}
	}

	result := httpResult{StatusCode: resp.StatusCode, Header: resp.Header.Clone(), Body: bodyBytes}
	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if strings.HasPrefix(contentType, "application/json") {
		var payload map[string]any
		if err := json.Unmarshal(bodyBytes, &payload); err == nil {
			result.JSON = payload
		}
	}
	return result
}

func percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	idx := int(math.Ceil((float64(p)/100.0)*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func maxDuration(sorted []float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	return sorted[len(sorted)-1]
}

func loginAndRefresh(t testing.TB, runtime *e2eharness.Runtime) (string, string, error) {
	loginResp := sendRequest(t, runtime, http.MethodPost, "/api/v1/auth/google", map[string]any{
		"token": e2eharness.ValidGoogleToken(),
	}, nil)
	if loginResp.StatusCode != fiber.StatusOK {
		return "", "", fmt.Errorf("login failed: status=%d body=%s", loginResp.StatusCode, string(loginResp.Body))
	}
	httpharness.RequireSuccessEnvelope(t, loginResp.JSON)

	accessToken := readSessionAccessToken(loginResp.JSON)
	refreshCookie := extractCookiePair(loginResp.Header, "recova_refresh_e2e")
	if strings.TrimSpace(refreshCookie) == "" {
		return "", "", fmt.Errorf("refresh cookie not found")
	}

	refreshResp := sendRequest(t, runtime, http.MethodPost, "/api/v1/auth/refresh", nil, map[string]string{"Cookie": refreshCookie})
	if refreshResp.StatusCode != fiber.StatusOK {
		return "", "", fmt.Errorf("refresh failed: status=%d body=%s", refreshResp.StatusCode, string(refreshResp.Body))
	}
	httpharness.RequireSuccessEnvelope(t, refreshResp.JSON)

	newAccess := readSessionAccessToken(refreshResp.JSON)
	if strings.TrimSpace(newAccess) != "" {
		accessToken = newAccess
	}
	if strings.TrimSpace(accessToken) == "" {
		return "", "", fmt.Errorf("access token empty")
	}

	return accessToken, refreshCookie, nil
}

func bearerHeaders(accessToken string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + strings.TrimSpace(accessToken),
	}
}

func prepareCommunityThread(runtime *e2eharness.Runtime, accessToken string) (string, error) {
	createPostResp := sendRequest(nil, runtime, http.MethodPost, "/api/v1/community", map[string]any{
		"content":  "load smoke community post",
		"category": "motivation",
	}, bearerHeaders(accessToken))
	if createPostResp.StatusCode != fiber.StatusCreated {
		return "", fmt.Errorf("create post failed: status=%d body=%s", createPostResp.StatusCode, string(createPostResp.Body))
	}
	postID := readNestedString(createPostResp.JSON, "data", "id")
	if strings.TrimSpace(postID) == "" {
		return "", fmt.Errorf("post id empty")
	}

	createCommentResp := sendRequest(nil, runtime, http.MethodPost, "/api/v1/community/"+postID+"/comments", map[string]any{
		"content": "load smoke root comment",
	}, bearerHeaders(accessToken))
	if createCommentResp.StatusCode != fiber.StatusCreated {
		return "", fmt.Errorf("create comment failed: status=%d body=%s", createCommentResp.StatusCode, string(createCommentResp.Body))
	}
	commentID := readNestedString(createCommentResp.JSON, "data", "id")
	if strings.TrimSpace(commentID) == "" {
		return "", fmt.Errorf("comment id empty")
	}

	createReplyResp := sendRequest(nil, runtime, http.MethodPost, "/api/v1/community/"+postID+"/comments/"+commentID+"/replies", map[string]any{
		"content": "load smoke reply comment",
	}, bearerHeaders(accessToken))
	if createReplyResp.StatusCode != fiber.StatusCreated {
		return "", fmt.Errorf("create reply failed: status=%d body=%s", createReplyResp.StatusCode, string(createReplyResp.Body))
	}

	return postID, nil
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
	value, _ := session["accessToken"].(string)
	return strings.TrimSpace(value)
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
