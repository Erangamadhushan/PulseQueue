package jobs

import "testing"

func TestJobStatusIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{
			name:   "pending is valid",
			status: JobStatusPending,
			want:   true,
		},
		{
			name:   "running is valid",
			status: JobStatusRunning,
			want:   true,
		},
		{
			name:   "completed is valid",
			status: JobStatusCompleted,
			want:   true,
		},
		{
			name:   "failed is valid",
			status: JobStatusFailed,
			want:   true,
		},
		{
			name:   "cancelled is valid",
			status: JobStatusCancelled,
			want:   true,
		},
		{
			name:   "unknown is invalid",
			status: JobStatus("unknown"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()

			if got != tt.want {
				t.Errorf(
					"IsValid() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestJobStatusIsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status JobStatus
		want   bool
	}{
		{
			name:   "pending is not terminal",
			status: JobStatusPending,
			want:   false,
		},
		{
			name:   "running is not terminal",
			status: JobStatusRunning,
			want:   false,
		},
		{
			name:   "completed is terminal",
			status: JobStatusCompleted,
			want:   true,
		},
		{
			name:   "failed is terminal",
			status: JobStatusFailed,
			want:   true,
		},
		{
			name:   "cancelled is terminal",
			status: JobStatusCancelled,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsTerminal()

			if got != tt.want {
				t.Errorf(
					"IsTerminal() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestJobStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from JobStatus
		to   JobStatus
		want bool
	}{
		{
			name: "pending to running",
			from: JobStatusPending,
			to:   JobStatusRunning,
			want: true,
		},
		{
			name: "pending to cancelled",
			from: JobStatusPending,
			to:   JobStatusCancelled,
			want: true,
		},
		{
			name: "pending cannot complete directly",
			from: JobStatusPending,
			to:   JobStatusCompleted,
			want: false,
		},
		{
			name: "running to completed",
			from: JobStatusRunning,
			to:   JobStatusCompleted,
			want: true,
		},
		{
			name: "running to pending for retry",
			from: JobStatusRunning,
			to:   JobStatusPending,
			want: true,
		},
		{
			name: "running to failed",
			from: JobStatusRunning,
			to:   JobStatusFailed,
			want: true,
		},
		{
			name: "completed is terminal",
			from: JobStatusCompleted,
			to:   JobStatusPending,
			want: false,
		},
		{
			name: "failed is terminal",
			from: JobStatusFailed,
			to:   JobStatusPending,
			want: false,
		},
		{
			name: "cancelled is terminal",
			from: JobStatusCancelled,
			to:   JobStatusPending,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)

			if got != tt.want {
				t.Errorf(
					"CanTransitionTo(%q) = %v, want %v",
					tt.to,
					got,
					tt.want,
				)
			}
		})
	}
}
