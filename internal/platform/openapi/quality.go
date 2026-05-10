package openapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ValidateProductionReadiness enforces non-negotiable OpenAPI quality gates.
func ValidateProductionReadiness(doc *openapi3.T) error {
	if doc == nil {
		return fmt.Errorf("openapi document is nil")
	}
	if doc.Paths == nil {
		return fmt.Errorf("openapi document has no paths")
	}

	issues := make([]string, 0, 32)
	seenOperationIDs := map[string]string{}

	for _, path := range doc.Paths.InMatchingOrder() {
		item := doc.Paths.Value(path)
		if item == nil {
			continue
		}

		operations := item.Operations()
		methods := make([]string, 0, len(operations))
		for method := range operations {
			methods = append(methods, method)
		}
		sort.Strings(methods)

		for _, method := range methods {
			op := operations[method]
			if op == nil {
				continue
			}

			opKey := strings.ToUpper(method) + " " + path
			opID := strings.TrimSpace(op.OperationID)
			if opID == "" {
				issues = append(issues, fmt.Sprintf("%s missing operationId", opKey))
			} else if prev, exists := seenOperationIDs[opID]; exists {
				issues = append(issues, fmt.Sprintf("%s duplicate operationId=%q already used by %s", opKey, opID, prev))
			} else {
				seenOperationIDs[opID] = opKey
			}

			if op.Responses == nil || op.Responses.Len() == 0 {
				issues = append(issues, fmt.Sprintf("%s has no responses", opKey))
				continue
			}

			responseCodes := op.Responses.Keys()
			sort.Strings(responseCodes)
			for _, code := range responseCodes {
				ref := op.Responses.Value(code)
				if ref == nil || ref.Value == nil {
					continue
				}
				content := ref.Value.Content
				if content == nil {
					continue
				}

				mediaTypes := make([]string, 0, len(content))
				for mediaType := range content {
					mediaTypes = append(mediaTypes, mediaType)
				}
				sort.Strings(mediaTypes)
				for _, mediaType := range mediaTypes {
					if !strings.HasPrefix(strings.ToLower(mediaType), "application/json") {
						continue
					}

					mt := content[mediaType]
					if mt == nil {
						issues = append(issues, fmt.Sprintf("%s response %s %s mediaType is nil", opKey, code, mediaType))
						continue
					}
					if mt.Schema == nil {
						issues = append(issues, fmt.Sprintf("%s response %s %s missing schema", opKey, code, mediaType))
					}
					if mt.Example == nil && len(mt.Examples) == 0 {
						issues = append(issues, fmt.Sprintf("%s response %s %s missing examples", opKey, code, mediaType))
						continue
					}

					if hasTemplatePlaceholder(mt.Example) {
						issues = append(issues, fmt.Sprintf("%s response %s %s uses placeholder in example", opKey, code, mediaType))
					}
					for name, exRef := range mt.Examples {
						if exRef == nil || exRef.Value == nil {
							continue
						}
						if hasTemplatePlaceholder(exRef.Value.Value) {
							issues = append(issues, fmt.Sprintf("%s response %s %s example %q uses placeholder", opKey, code, mediaType, name))
						}
					}
				}
			}
		}
	}

	if len(issues) > 0 {
		return fmt.Errorf("openapi production-readiness failed:\n- %s", strings.Join(issues, "\n- "))
	}

	return nil
}

func hasTemplatePlaceholder(v any) bool {
	switch typed := v.(type) {
	case nil:
		return false
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "string" || normalized == "anything"
	case []any:
		for _, item := range typed {
			if hasTemplatePlaceholder(item) {
				return true
			}
		}
		return false
	case map[string]any:
		for k, item := range typed {
			if strings.EqualFold(strings.TrimSpace(k), "additionalProperty") {
				return true
			}
			if hasTemplatePlaceholder(item) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
