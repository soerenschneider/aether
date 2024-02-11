package carddav

import (
	"reflect"
	"testing"
	"time"
)

func Test_sortCards(t *testing.T) {
	type args struct {
		entries []Card
		now     time.Time
	}
	tests := []struct {
		name string
		args args
		want []Card
	}{
		{
			name: "one upcoming date, one late day",
			args: args{
				entries: []Card{
					{
						Name:        "a",
						Anniversary: time.Date(1980, 12, 5, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 11, 30, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "b",
					Anniversary: time.Date(2023, 11, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "a",
					Anniversary: time.Date(1980, 12, 5, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "same month",
			args: args{
				entries: []Card{
					{
						Name:        "a",
						Anniversary: time.Date(1980, 12, 5, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "a",
					Anniversary: time.Date(1980, 12, 5, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "both dates in the next year, correct order",
			args: args{
				entries: []Card{
					{
						Name:        "a",
						Anniversary: time.Date(1980, 1, 5, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 2, 30, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "a",
					Anniversary: time.Date(1980, 1, 5, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 2, 30, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "two in the next year, one left in the current year",
			args: args{
				entries: []Card{
					{
						Name:        "a",
						Anniversary: time.Date(1980, 1, 5, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 2, 30, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "c",
						Anniversary: time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "c",
					Anniversary: time.Date(2023, 12, 30, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "a",
					Anniversary: time.Date(1980, 1, 5, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 2, 30, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "all in the upcoming year",
			args: args{
				entries: []Card{
					{
						Name:        "a",
						Anniversary: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "c",
						Anniversary: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "b",
					Anniversary: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "c",
					Anniversary: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "a",
					Anniversary: time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "all in the previous month",
			args: args{
				entries: []Card{
					{
						Name:        "b",
						Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "c",
						Anniversary: time.Date(2023, 11, 3, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "c",
					Anniversary: time.Date(2023, 11, 3, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "two in the previous month, one upcoming",
			args: args{
				entries: []Card{
					{
						Name:        "b",
						Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "b",
						Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
					},
					{
						Name:        "c",
						Anniversary: time.Date(2023, 12, 5, 0, 0, 0, 0, time.UTC),
					},
				},
				now: time.Date(2023, 12, 2, 0, 0, 0, 0, time.UTC),
			},
			want: []Card{
				{
					Name:        "c",
					Anniversary: time.Date(2023, 12, 5, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
				},
				{
					Name:        "b",
					Anniversary: time.Date(2023, 11, 4, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortCards(tt.args.entries, tt.args.now)
			if !reflect.DeepEqual(tt.want, tt.args.entries) {
				t.Fatalf("wanted=%v, got=%v", tt.want, tt.args.entries)
			}
		})
	}
}
