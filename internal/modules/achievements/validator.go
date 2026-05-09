package achievements

import (
	"strings"

	"github.com/recova-app/backend-v2/internal/shared/errs"
)

const (
	categoryStreakMilestone        = "streak_milestone"
	categoryCheckinConsistency     = "checkin_consistency"
	categoryJournalConsistency     = "journal_consistency"
	categoryRelapseRecovery        = "relapse_recovery"
	categoryCommunityParticipation = "community_participation"
	categoryOnboardingCompletion   = "onboarding_completion"
)

var supportedCategories = map[string]struct{}{
	categoryStreakMilestone:        {},
	categoryCheckinConsistency:     {},
	categoryJournalConsistency:     {},
	categoryRelapseRecovery:        {},
	categoryCommunityParticipation: {},
	categoryOnboardingCompletion:   {},
}

// NormalizeCategoryQuery validates and normalizes optional achievement category query.
func NormalizeCategoryQuery(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(*raw))
	if normalized == "" {
		return nil, nil
	}

	if _, ok := supportedCategories[normalized]; !ok {
		return nil, errs.New(errs.CodeValidationError, "Kategori achievement tidak valid", []map[string]string{
			{"field": "category", "message": "Kategori achievement tidak didukung"},
		}, nil)
	}

	return &normalized, nil
}
