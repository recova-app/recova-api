package database

import (
	"testing"

	"gorm.io/gorm"
)

func TestWithTransaction_NilDatabase_ReturnsError(t *testing.T) {
	err := WithTransaction(nil, nil, func(_ *gorm.DB) error { return nil })
	if err == nil {
		t.Fatal("expected error for nil database handle")
	}
}

func TestWithTransaction_NilCallback_ReturnsError(t *testing.T) {
	err := WithTransaction(nil, &gorm.DB{}, nil)
	if err == nil {
		t.Fatal("expected error for nil callback")
	}
}
