package cached

import (
	"reflect"
	"testing"
	"time"
)

func Test_calculateNextRefreshInterval(t *testing.T) {
	tests := []struct {
		name            string
		now             time.Time
		refreshInterval time.Duration
		want            time.Time
	}{
		{
			name:            "Case 1: refresh interval before midnight",
			now:             time.Date(2025, time.February, 7, 8, 0, 0, 0, time.UTC), // 8:00 AM
			refreshInterval: 3 * time.Hour,                                           // Refresh in 3 hours (11:00 AM)
			want:            time.Date(2025, time.February, 7, 11, 0, 0, 0, time.UTC),
		},
		{
			name:            "Case 2: refresh interval after midnight (e.g., 9 hours later -> 5:00 PM)",
			now:             time.Date(2025, time.February, 7, 20, 0, 0, 0, time.UTC), // 8:00 AM
			refreshInterval: 9 * time.Hour,                                            // Refresh in 9 hours (5:00 PM)
			want:            time.Date(2025, time.February, 8, 0, 0, 0, 0, time.UTC),  // Midnight
		},
		{
			name:            "// Case 3: refresh interval exactly at midnight",
			now:             time.Date(2025, time.February, 7, 23, 59, 59, 0, time.UTC), // 11:59:59 PM
			refreshInterval: 1 * time.Second,                                            // Refresh in 1 second (midnight)
			want:            time.Date(2025, time.February, 8, 0, 0, 0, 0, time.UTC),    // Midnight
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateNextRefreshInterval(tt.refreshInterval, tt.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateNextRefreshInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}
