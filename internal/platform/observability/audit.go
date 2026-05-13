package observability

import (
	"log/slog"
	"strings"
)

type auditRoute struct {
	method string
	path   string
	action string
}

var auditedRoutes = []auditRoute{
	{method: "POST", path: "/api/v1/auth/google", action: "auth.login"},
	{method: "POST", path: "/api/v1/auth/refresh", action: "auth.refresh"},
	{method: "POST", path: "/api/v1/auth/logout", action: "auth.logout"},
	{method: "POST", path: "/api/v1/auth/onboarding", action: "auth.onboarding"},
	{method: "PUT", path: "/api/v1/users/settings", action: "users.settings.update"},
	{method: "DELETE", path: "/api/v1/users/me/reset-data", action: "users.data.reset"},
	{method: "PUT", path: "/api/v1/ai/persona-preferences", action: "ai.persona.preference.update"},
	{method: "POST", path: "/api/v1/community/:post_id/comments/:comment_id/replies", action: "community.comment.reply"},
}

func auditAction(method string, path string) (string, bool) {
	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedPath := normalizeRoute(path)
	for _, route := range auditedRoutes {
		if route.method == normalizedMethod && route.path == normalizedPath {
			return route.action, true
		}
	}
	return "", false
}

// RecordAuditEvent logs and tracks audit event for sensitive routes.
func RecordAuditEvent(logger *slog.Logger, recorder *Recorder, ctx RequestContext) {
	action, ok := auditAction(ctx.Method, ctx.RoutePath)
	if !ok {
		return
	}

	result := "failed"
	if ctx.StatusCode >= 200 && ctx.StatusCode < 400 {
		result = "succeeded"
	}

	if recorder != nil {
		recorder.RecordAuditEvent(action, result)
	}

	if logger != nil {
		logger.Info("audit event",
			"action", action,
			"result", result,
			"requestId", strings.TrimSpace(ctx.RequestID),
			"userId", strings.TrimSpace(ctx.UserID),
			"status", ctx.StatusCode,
			"method", strings.ToUpper(strings.TrimSpace(ctx.Method)),
			"path", normalizeRoute(ctx.RoutePath),
		)
	}
}
