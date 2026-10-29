package pkg

import (
	"testing"
	"time"
)

func TestDurationToString(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{0, "0 days"},
		{24 * time.Hour, "1 day"},
		{3 * 24 * time.Hour, "3 days"},
		{31 * 24 * time.Hour, "1 month, 1 day"},
		{(365 + 61) * 24 * time.Hour, "1 year, 2 months, 1 day"},
	}

	for _, tt := range tests {
		if got := DurationToString(tt.duration); got != tt.want {
			t.Errorf("DurationToString(%v) = %q, want %q", tt.duration, got, tt.want)
		}
	}
}
