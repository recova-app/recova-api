package contract

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const modulesRootPath = "../../internal/modules"

var expectedModules = []string{
	"achievements",
	"ai",
	"auth",
	"community",
	"content",
	"education",
	"journals",
	"routine",
	"users",
}

var coreFiles = []string{
	"doc.go",
	"dto.go",
	"handler.go",
	"repository.go",
	"route.go",
	"service.go",
	"validator.go",
}

type trackedGap struct {
	Status string
	Owner  string
	Reason string
}

var trackedCoreGaps = map[string]trackedGap{
	gapKey("content", "validator.go"): {
		Status: "accepted-exception",
		Owner:  "content-owner",
		Reason: "modul read-only tanpa payload tulis; validasi request custom belum diperlukan",
	},
	gapKey("education", "validator.go"): {
		Status: "accepted-exception",
		Owner:  "education-owner",
		Reason: "modul read-only tanpa payload tulis; validasi request custom belum diperlukan",
	},
}

var trackedCompanionGaps = map[string]trackedGap{
	gapKey("content", "repository_test.go|repository_integration_test.go"): {
		Status: "accepted-exception",
		Owner:  "content-owner",
		Reason: "coverage repository content ditopang seed integration test lintas modul",
	},
}

func TestContract_ModuleStructureConsistency_Baseline(t *testing.T) {
	modules := readModules(t)
	assertModuleSet(t, modules)

	actualCoreGaps := map[string]struct{}{}
	actualCompanionGaps := map[string]struct{}{}

	for _, module := range modules {
		modulePath := filepath.Join(modulesRootPath, module)

		for _, core := range coreFiles {
			if fileExists(filepath.Join(modulePath, core)) {
				continue
			}
			actualCoreGaps[gapKey(module, core)] = struct{}{}
		}

		if fileExists(filepath.Join(modulePath, "route.go")) && !fileExists(filepath.Join(modulePath, "route_test.go")) {
			actualCompanionGaps[gapKey(module, "route_test.go")] = struct{}{}
		}

		if fileExists(filepath.Join(modulePath, "service.go")) && !fileExists(filepath.Join(modulePath, "service_test.go")) {
			actualCompanionGaps[gapKey(module, "service_test.go")] = struct{}{}
		}

		if fileExists(filepath.Join(modulePath, "validator.go")) && !fileExists(filepath.Join(modulePath, "validator_test.go")) {
			actualCompanionGaps[gapKey(module, "validator_test.go")] = struct{}{}
		}

		if fileExists(filepath.Join(modulePath, "repository.go")) &&
			!fileExists(filepath.Join(modulePath, "repository_test.go")) &&
			!fileExists(filepath.Join(modulePath, "repository_integration_test.go")) {
			actualCompanionGaps[gapKey(module, "repository_test.go|repository_integration_test.go")] = struct{}{}
		}
	}

	assertTrackedGapMatch(t, "core", actualCoreGaps, trackedCoreGaps)
	assertTrackedGapMatch(t, "companion-test", actualCompanionGaps, trackedCompanionGaps)
}

func TestContract_ModuleLayerAndRouteConsistency_Baseline(t *testing.T) {
	modules := readModules(t)

	routeDefaultName := regexp.MustCompile(`func\s+RegisterRoutes\s*\(`)
	routeAuthName := regexp.MustCompile(`func\s+RegisterCoreRoutes\s*\(`)
	routeUsersName := regexp.MustCompile(`func\s+RegisterUserRoutes\s*\(`)
	routeUsersOnboardingName := regexp.MustCompile(`func\s+RegisterOnboardingRoute\s*\(`)

	for _, module := range modules {
		modulePath := filepath.Join(modulesRootPath, module)

		handlerPath := filepath.Join(modulePath, "handler.go")
		handlerContent := readFile(t, handlerPath)
		if strings.Contains(handlerContent, "internal/platform/database") || strings.Contains(handlerContent, "gorm.io/gorm") {
			t.Fatalf("handler must not import database adapter directly: %s", handlerPath)
		}
		if !strings.Contains(handlerContent, "response.Success(") {
			t.Fatalf("handler should use shared success envelope: %s", handlerPath)
		}

		servicePath := filepath.Join(modulePath, "service.go")
		serviceContent := readFile(t, servicePath)
		if module != "auth" && (strings.Contains(serviceContent, "fiber.Ctx") || strings.Contains(serviceContent, "gofiber/fiber")) {
			t.Fatalf("service must stay HTTP-agnostic: %s", servicePath)
		}

		repositoryPath := filepath.Join(modulePath, "repository.go")
		repositoryContent := readFile(t, repositoryPath)
		if strings.Contains(repositoryContent, "internal/shared/response") || strings.Contains(repositoryContent, "gofiber/fiber") {
			t.Fatalf("repository must not format HTTP response: %s", repositoryPath)
		}

		routePath := filepath.Join(modulePath, "route.go")
		routeContent := readFile(t, routePath)
		switch module {
		case "auth":
			if !routeAuthName.MatchString(routeContent) {
				t.Fatalf("auth route registration should use RegisterCoreRoutes: %s", routePath)
			}
		case "users":
			if !routeUsersName.MatchString(routeContent) || !routeUsersOnboardingName.MatchString(routeContent) {
				t.Fatalf("users route registration should expose RegisterUserRoutes and RegisterOnboardingRoute: %s", routePath)
			}
		default:
			if !routeDefaultName.MatchString(routeContent) {
				t.Fatalf("route registration should use RegisterRoutes: %s", routePath)
			}
		}

		if module == "auth" {
			continue
		}
		assertProtectedRouteAuthGuard(t, module, routePath, routeContent)
	}
}

func assertProtectedRouteAuthGuard(t testing.TB, module string, routePath string, content string) {
	t.Helper()

	allowAuthGuardAlias := strings.Contains(content, "authGuard := authmodule.RequireAuth(")
	lines := strings.Split(content, "\n")
	registeredRouteLines := 0
	authGuardedRouteLines := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "router.") {
			continue
		}
		if strings.HasPrefix(trimmed, "router.Use(") {
			continue
		}
		if !strings.Contains(trimmed, "(") || !strings.Contains(trimmed, ",") {
			continue
		}
		registeredRouteLines++
		if strings.Contains(trimmed, "RequireAuth(") {
			authGuardedRouteLines++
			continue
		}
		if allowAuthGuardAlias && strings.Contains(trimmed, "authGuard,") {
			authGuardedRouteLines++
		}
	}

	if registeredRouteLines == 0 {
		t.Fatalf("no registered routes detected in module %s: %s", module, routePath)
	}

	if registeredRouteLines != authGuardedRouteLines {
		t.Fatalf(
			"module %s has non-auth-guarded routes (%d/%d) in %s",
			module,
			authGuardedRouteLines,
			registeredRouteLines,
			routePath,
		)
	}
}

func assertTrackedGapMatch(
	t testing.TB,
	label string,
	actual map[string]struct{},
	tracked map[string]trackedGap,
) {
	t.Helper()

	for key := range actual {
		if _, ok := tracked[key]; ok {
			continue
		}
		t.Fatalf("untracked %s gap detected: %s", label, key)
	}

	for key, info := range tracked {
		if _, ok := actual[key]; ok {
			if strings.TrimSpace(info.Status) == "" || strings.TrimSpace(info.Owner) == "" || strings.TrimSpace(info.Reason) == "" {
				t.Fatalf("tracked %s gap metadata incomplete: %s", label, key)
			}
			if info.Status != "accepted-exception" && info.Status != "needs-follow-up" {
				t.Fatalf("tracked %s gap status invalid: %s (%s)", label, key, info.Status)
			}
			continue
		}
		t.Fatalf("stale tracked %s gap, remove from registry: %s", label, key)
	}
}

func assertModuleSet(t testing.TB, modules []string) {
	t.Helper()

	want := append([]string(nil), expectedModules...)
	slices.Sort(want)

	got := append([]string(nil), modules...)
	slices.Sort(got)

	if len(got) != len(want) {
		t.Fatalf("module count mismatch: got=%v want=%v", got, want)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("module set mismatch: got=%v want=%v", got, want)
		}
	}
}

func readModules(t testing.TB) []string {
	t.Helper()

	entries, err := os.ReadDir(modulesRootPath)
	if err != nil {
		t.Fatalf("read modules root: %v", err)
	}

	modules := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modules = append(modules, entry.Name())
	}
	return modules
}

func readFile(t testing.TB, path string) string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(raw)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func gapKey(module string, file string) string {
	return fmt.Sprintf("%s:%s", module, file)
}
