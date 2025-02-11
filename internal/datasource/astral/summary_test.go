package astral

import (
	"reflect"
	"testing"
	"time"
)

func Test_getSummary(t *testing.T) {
	type args struct {
		data AstralData
		now  time.Time
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "happy case",
			args: args{
				data: AstralData{
					Sunrise:           time.Date(2025, 2, 10, 6, 30, 0, 0, time.UTC),
					Sunset:            time.Date(2025, 2, 10, 18, 15, 0, 0, time.UTC),
					BlueHourRising:    TimeDuration{Start: time.Date(2025, 2, 10, 6, 0, 0, 0, time.UTC), End: time.Date(2025, 2, 10, 6, 20, 0, 0, time.UTC)},
					BlueHourSetting:   TimeDuration{Start: time.Date(2025, 2, 10, 17, 50, 0, 0, time.UTC), End: time.Date(2025, 2, 10, 18, 10, 0, 0, time.UTC)},
					GoldenHourRising:  TimeDuration{Start: time.Date(2025, 2, 10, 7, 0, 0, 0, time.UTC), End: time.Date(2025, 2, 10, 7, 30, 0, 0, time.UTC)},
					GoldenHourSetting: TimeDuration{Start: time.Date(2025, 2, 10, 17, 0, 0, 0, time.UTC), End: time.Date(2025, 2, 10, 17, 30, 0, 0, time.UTC)},
				},
				now: time.Date(2025, 2, 10, 12, 15, 0, 0, time.UTC),
			},
			want: []string{
				"☀️Sun: 06:30 - 18:15",
				"🌇 Golden Hour ⬇️️ 17:00 - 17:30, 🌌 Blue Hour ⬇️ 17:50 - 18:10",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSummary(tt.args.data, tt.args.now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSummary() = %v, want %v", got, tt.want)
			}
		})
	}
}
