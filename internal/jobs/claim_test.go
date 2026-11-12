package jobs

import (
	"errors"
	"testing"
	"time"
)

func TestValidateClaimInputDefaultsToDefaultQueue(t *testing.T) {
	queues, err := validateClaimInput(ClaimJobInput{
		WorkerID:      "worker-1",
		LeaseDuration: time.Minute,
	})
	if err != nil {
		t.Fatalf("validateClaimInput() error = %v", err)
	}
	if len(queues) != 1 || queues[0] != DefaultQueue {
		t.Fatalf("queues = %v, want [%q]", queues, DefaultQueue)
	}
}

func TestValidateClaimInputNormalizesQueues(t *testing.T) {
	queues, err := validateClaimInput(ClaimJobInput{
		WorkerID:      " worker-1 ",
		LeaseDuration: time.Minute,
		Queues:        []string{" emails ", "emails", "default"},
	})
	if err != nil {
		t.Fatalf("validateClaimInput() error = %v", err)
	}
	want := []string{"emails", "default"}
	if len(queues) != len(want) || queues[0] != want[0] || queues[1] != want[1] {
		t.Fatalf("queues = %v, want %v", queues, want)
	}
}

func TestValidateClaimInputRejectsInvalidValues(t *testing.T) {
	tests := []ClaimJobInput{
		{LeaseDuration: time.Minute, Queues: []string{DefaultQueue}},
		{WorkerID: "worker-1", Queues: []string{DefaultQueue}},
		{WorkerID: "worker-1", LeaseDuration: time.Minute, Queues: []string{" "}},
	}

	for _, input := range tests {
		if _, err := validateClaimInput(input); !errors.Is(err, ErrInvalidClaimInput) {
			t.Errorf("validateClaimInput(%+v) error = %v, want ErrInvalidClaimInput", input, err)
		}
	}
}
