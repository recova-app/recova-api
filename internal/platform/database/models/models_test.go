package models

import "testing"

func TestTableNames_AreStable(t *testing.T) {
	if (AuthRefreshToken{}).TableName() != "auth_refresh_tokens" {
		t.Fatalf("unexpected auth refresh tokens table name: %s", (AuthRefreshToken{}).TableName())
	}
	if (UserAchievementProgress{}).TableName() != "user_achievement_progress" {
		t.Fatalf("unexpected user achievement progress table name: %s", (UserAchievementProgress{}).TableName())
	}
	if (UserAIPersonaPreference{}).TableName() != "user_ai_persona_preferences" {
		t.Fatalf("unexpected user ai persona preferences table name: %s", (UserAIPersonaPreference{}).TableName())
	}
}

func TestStringPtr_ReturnsPointer(t *testing.T) {
	value := "hello"
	ptr := StringPtr(value)
	if ptr == nil || *ptr != value {
		t.Fatalf("unexpected pointer value: %#v", ptr)
	}
}
