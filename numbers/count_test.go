package numbers

import (
	"errors"
	"github.com/semsen1/aggregation-go/common"
	"testing"
)

func TestCountIf(t *testing.T) {
	got, err := CountIf([]int{1, 2, 3, 4, 5}, func(v int) bool { return v%2 == 1 })
	if err != nil || got != 3 {
		t.Fatalf("CountIf = %v, %v; want 3", got, err)
	}
	got, err = CountIf([]string{"go", "rust", "c"}, func(s string) bool { return len(s) == 2 })
	if err != nil || got != 1 {
		t.Fatalf("CountIf strings = %v, %v; want 1", got, err)
	}
}
func TestCountIfEmptyAndNil(t *testing.T) {
	got, err := CountIf[int](nil, func(int) bool { return true })
	if err != nil || got != 0 {
		t.Fatalf("empty = %v, %v; want 0, nil", got, err)
	}
	if _, err := CountIf([]int{1}, nil); !errors.Is(err, common.ErrNilFunc) {
		t.Fatalf("err = %v, want ErrNilFunc", err)
	}
}
