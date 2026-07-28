package models

import (
	"testing"
	"time"
)

// hoursAgo is relative to a caller-supplied instant. Deriving it from
// time.Now() independently would make the exact-boundary case flaky, since the
// two calls land microseconds apart.
func hoursAgo(now time.Time, h int) *time.Time {
	t := now.Add(-time.Duration(h) * time.Hour)
	return &t
}

func TestBackupScheduleDue(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name string
		s    BackupSchedule
		want bool
	}{
		{
			// Enabling a schedule should visibly do something rather than going
			// quiet for a full interval.
			name: "never run yet runs immediately",
			s:    BackupSchedule{Enabled: true, IntervalHours: 24},
			want: true,
		},
		{
			name: "disabled never runs even when overdue",
			s:    BackupSchedule{Enabled: false, IntervalHours: 1, LastRunAt: hoursAgo(now, 100)},
			want: false,
		},
		{
			name: "interval not yet elapsed",
			s:    BackupSchedule{Enabled: true, IntervalHours: 24, LastRunAt: hoursAgo(now, 1)},
			want: false,
		},
		{
			name: "interval elapsed",
			s:    BackupSchedule{Enabled: true, IntervalHours: 24, LastRunAt: hoursAgo(now, 25)},
			want: true,
		},
		{
			name: "exactly at the interval boundary is due",
			s:    BackupSchedule{Enabled: true, IntervalHours: 24, LastRunAt: hoursAgo(now, 24)},
			want: true,
		},
		{
			// Guards against a zero/negative interval turning into a hot loop
			// that backs up continuously.
			name: "zero interval never runs",
			s:    BackupSchedule{Enabled: true, IntervalHours: 0, LastRunAt: hoursAgo(now, 100)},
			want: false,
		},
		{
			name: "negative interval never runs",
			s:    BackupSchedule{Enabled: true, IntervalHours: -5, LastRunAt: hoursAgo(now, 100)},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.s.Due(now); got != tc.want {
				t.Fatalf("Due() = %v, want %v", got, tc.want)
			}
		})
	}
}
