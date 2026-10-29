package cached

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/soerenschneider/aether/internal"
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

type fakeDatasource struct {
	data []*internal.Data
	// errors returned instead of the data at the same position
	errs  []error
	calls int
}

func (f *fakeDatasource) Name() string {
	return "fake"
}

func (f *fakeDatasource) GetData(_ context.Context) (*internal.Data, error) {
	i := min(f.calls, len(f.data)-1)
	f.calls++
	if i < len(f.errs) && f.errs[i] != nil {
		return nil, f.errs[i]
	}
	return f.data[i], nil
}

func TestCachedDatasource_incompleteData(t *testing.T) {
	incomplete := &internal.Data{Summary: []string{"private only"}, Incomplete: true}
	complete := &internal.Data{Summary: []string{"private and family"}}
	fake := &fakeDatasource{data: []*internal.Data{incomplete, complete}}

	cache, err := New(fake, WithRefreshInterval(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := cache.GetData(context.Background())
	if err != nil || data != incomplete {
		t.Fatalf("expected the incomplete data to be shown, got %v, %v", data, err)
	}
	if until := time.Until(cache.nextRefresh); until > incompleteRetryInterval {
		t.Errorf("expected incomplete data to be fetched again within %v, but it's kept for %v", incompleteRetryInterval, until)
	}

	// cached until the retry is due
	if data, _ := cache.GetData(context.Background()); data != incomplete || fake.calls != 1 {
		t.Errorf("expected the cached data without fetching again, got %v after %d calls", data, fake.calls)
	}

	cache.nextRefresh = time.Now().Add(-time.Second)
	if data, _ := cache.GetData(context.Background()); data != complete {
		t.Fatalf("expected the complete data after the retry, got %v", data)
	}
	// the cache always refreshes at midnight, which may be closer than the retry interval
	nearMidnight := time.Until(calculateNextRefreshInterval(24*time.Hour, time.Now())) <= incompleteRetryInterval
	if until := time.Until(cache.nextRefresh); until <= incompleteRetryInterval && !nearMidnight {
		t.Errorf("expected complete data to be kept for the regular interval, but it's only kept for %v", until)
	}
}

func TestCachedDatasource_outdatedData(t *testing.T) {
	boom := errors.New("boom")
	first := &internal.Data{Summary: []string{"first"}}
	second := &internal.Data{Summary: []string{"second"}}
	fake := &fakeDatasource{data: []*internal.Data{first, nil, second}, errs: []error{nil, boom, nil}}

	cache, err := New(fake, WithRefreshInterval(30*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	data, err := cache.GetData(context.Background())
	if err != nil || data != first || data.Outdated || data.FetchedAt.IsZero() {
		t.Fatalf("unexpected first result: %+v, %v", data, err)
	}

	// the update fails: the last data is shown, marked as outdated
	cache.nextRefresh = time.Now().Add(-time.Second)
	data, err = cache.GetData(context.Background())
	if err != nil {
		t.Fatalf("expected the last data instead of an error, got %v", err)
	}
	if !data.Outdated || data.Summary[0] != "first" || !data.FetchedAt.Equal(first.FetchedAt) {
		t.Errorf("expected the first data marked as outdated, got %+v", data)
	}
	if first.Outdated {
		t.Error("the cached data must not be modified")
	}

	// the next call tries again and succeeds
	data, err = cache.GetData(context.Background())
	if err != nil || data != second || data.Outdated {
		t.Errorf("expected the second data after a successful retry, got %+v, %v", data, err)
	}
	if fake.calls != 3 {
		t.Errorf("expected 3 calls, got %d", fake.calls)
	}
}

func TestCachedDatasource_outdatedDataFromYesterday(t *testing.T) {
	boom := errors.New("boom")
	fake := &fakeDatasource{data: []*internal.Data{nil}, errs: []error{boom}}
	cache, _ := New(fake)

	// no data at all
	if _, err := cache.GetData(context.Background()); !errors.Is(err, boom) {
		t.Errorf("expected the error without earlier data, got %v", err)
	}

	// rendered data contains relative dates, so data from yesterday is not shown
	cache.data = &internal.Data{Summary: []string{"yesterday"}, FetchedAt: time.Now().AddDate(0, 0, -1)}
	if _, err := cache.GetData(context.Background()); !errors.Is(err, boom) {
		t.Errorf("expected the error instead of data from yesterday, got %v", err)
	}
}
