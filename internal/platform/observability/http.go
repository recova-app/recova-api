package observability

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	authmodule "github.com/recova-app/backend-v2/internal/modules/auth"
	"github.com/recova-app/backend-v2/internal/shared/errs"
)

// RequestHooks stores optional callbacks for request-level telemetry.
type RequestHooks struct {
	OnComplete func(ctx RequestContext)
}

// RequestContext captures stable request telemetry attributes.
type RequestContext struct {
	RequestID  string
	Method     string
	RoutePath  string
	StatusCode int
	Latency    time.Duration
	UserID     string
}

// NewRequestTelemetryMiddleware builds one middleware for logging, metrics, and audit.
func NewRequestTelemetryMiddleware(
	logger *slog.Logger,
	recorder *Recorder,
	requestIDHeader string,
	hooks RequestHooks,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		started := time.Now()
		err := c.Next()
		latency := time.Since(started)

		statusCode := c.Response().StatusCode()
		if err != nil {
			mapped := errs.Map(err)
			statusCode = mapped.Status
		}
		method := strings.TrimSpace(c.Method())
		routePath := currentRoutePath(c)

		reqID := strings.TrimSpace(requestid.FromContext(c))
		if reqID == "" {
			reqID = strings.TrimSpace(c.Get(requestIDHeader))
		}

		userID := ""
		if principal, ok := authmodule.PrincipalFromContext(c); ok {
			userID = strings.TrimSpace(principal.UserID)
		}

		if recorder != nil {
			recorder.RecordHTTPRequest(method, routePath, statusCode, latency)
		}

		if err != nil {
			logger.Error("request completed",
				"requestId", reqID,
				"method", method,
				"path", routePath,
				"status", statusCode,
				"latencyMs", latency.Milliseconds(),
				"userId", userID,
				"error", err,
			)
		} else {
			logger.Info("request completed",
				"requestId", reqID,
				"method", method,
				"path", routePath,
				"status", statusCode,
				"latencyMs", latency.Milliseconds(),
				"userId", userID,
			)
		}

		if hooks.OnComplete != nil {
			hooks.OnComplete(RequestContext{
				RequestID:  reqID,
				Method:     method,
				RoutePath:  routePath,
				StatusCode: statusCode,
				Latency:    latency,
				UserID:     userID,
			})
		}

		return err
	}
}

func currentRoutePath(c fiber.Ctx) string {
	if route := c.Route(); route != nil {
		if normalized := strings.TrimSpace(route.Path); normalized != "" {
			if strings.HasPrefix(normalized, "/") {
				return normalized
			}
			return "/" + normalized
		}
	}
	return normalizeRoute(c.Path())
}
