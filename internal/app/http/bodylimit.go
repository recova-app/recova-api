package http

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var bodyLimitUnitMultiplier = map[string]int64{
	"b":  1,
	"kb": 1024,
	"mb": 1024 * 1024,
	"gb": 1024 * 1024 * 1024,
}

var bodyLimitUnitsByPriority = []string{"gb", "mb", "kb", "b"}

func parseBodyLimitBytes(value string) (int, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return 0, fmt.Errorf("value is empty")
	}

	unit := "b"
	number := normalized

	for _, suffix := range bodyLimitUnitsByPriority {
		if strings.HasSuffix(normalized, suffix) {
			unit = suffix
			number = strings.TrimSpace(strings.TrimSuffix(normalized, suffix))
			break
		}
	}

	if number == "" {
		return 0, fmt.Errorf("body limit number is missing")
	}

	parsedNumber, err := strconv.ParseInt(number, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid body limit number: %w", err)
	}

	if parsedNumber <= 0 {
		return 0, fmt.Errorf("body limit must be greater than 0")
	}

	multiplier, ok := bodyLimitUnitMultiplier[unit]
	if !ok {
		return 0, fmt.Errorf("unsupported body limit unit: %s", unit)
	}

	if parsedNumber > math.MaxInt64/multiplier {
		return 0, fmt.Errorf("body limit exceeds integer range")
	}

	total := parsedNumber * multiplier
	if total > int64(maxInt()) {
		return 0, fmt.Errorf("body limit exceeds integer range")
	}

	return int(total), nil
}

func maxInt() int {
	return int(^uint(0) >> 1)
}
