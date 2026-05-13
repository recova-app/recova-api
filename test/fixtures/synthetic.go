// Package fixtures contains synthetic non-secret fixtures for tests.
package fixtures

import "time"

// UserFixture models synthetic user identity fixture.
type UserFixture struct {
	ID            string
	GoogleID      string
	Email         string
	DisplayName   string
	RecoveryFocus string
}

// JournalFixture models synthetic journal content fixture.
type JournalFixture struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	CreatedAt time.Time
}

// RoutineCheckInFixture models synthetic check-in fixture.
type RoutineCheckInFixture struct {
	ID      string
	UserID  string
	Mood    string
	Note    string
	DayDate string
}

// SyntheticUser returns deterministic non-secret user fixture.
func SyntheticUser() UserFixture {
	return UserFixture{
		ID:            "00000000-0000-0000-0000-000000000001",
		GoogleID:      "google-sandbox-user-1",
		Email:         "synthetic.user1@example.test",
		DisplayName:   "Test User One",
		RecoveryFocus: "Building healthy routines",
	}
}

// SyntheticJournal returns deterministic non-secret journal fixture.
func SyntheticJournal() JournalFixture {
	return JournalFixture{
		ID:        "10000000-0000-0000-0000-000000000001",
		UserID:    SyntheticUser().ID,
		Title:     "First Day Reflection",
		Body:      "Today I successfully kept my check-in commitment.",
		CreatedAt: time.Date(2026, time.May, 8, 7, 0, 0, 0, time.UTC),
	}
}

// SyntheticRoutineCheckIn returns deterministic non-secret routine fixture.
func SyntheticRoutineCheckIn() RoutineCheckInFixture {
	return RoutineCheckInFixture{
		ID:      "20000000-0000-0000-0000-000000000001",
		UserID:  SyntheticUser().ID,
		Mood:    "motivated",
		Note:    "I focus on small, consistent steps.",
		DayDate: "2026-05-08",
	}
}
