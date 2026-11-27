package numbers

import (
	"errors"
	"github.com/semsen1/aggregation-go/common"
	"math"
	"testing"
)

func TestMinMax(t *testing.T) {
	s := []int{3, -1, 7, 0}
	if got, err := Min(s); err != nil || got != -1 {
		t.Fatalf("Min = %v, %v", got, err)
	}
	if got, err := Max(s); err != nil || got != 7 {
		t.Fatalf("Max = %v, %v", got, err)
	}
	if got, err := Min([]float64{2.5}); err != nil || got != 2.5 {
		t.Fatalf("single element Min = %v, %v", got, err)
	}
}
func TestMinMaxEmpty(t *testing.T) {
	if _, err := Min([]int(nil)); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("Min: err = %v, want ErrEmpty", err)
	}
	if _, err := Max([]float64{}); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("Max: err = %v, want ErrEmpty", err)
	}
}
func TestMinMaxNaNDoesNotDependOnPosition(t *testing.T) {
	nan := math.NaN()
	for _, s := range [][]float64{{nan, 1, 2}, {1, nan, 2}, {1, 2, nan}} {
		got, err := Min(s)
		if err != nil || !math.IsNaN(got) {
			t.Fatalf("Min(%v) = %v, %v; want NaN, nil", s, got, err)
		}
		got, err = Max(s)
		if err != nil || !math.IsNaN(got) {
			t.Fatalf("Max(%v) = %v, %v; want NaN, nil", s, got, err)
		}
	}
}
func TestMinMaxNaNSkip(t *testing.T) {
	nan := math.NaN()
	got, err := Min([]float64{nan, 3, 1, nan}, common.WithNaNPolicy(common.NaNSkip))
	if err != nil || got != 1 {
		t.Fatalf("Min = %v, %v; want 1", got, err)
	}
	if _, err := Min([]float64{nan}, common.WithNaNPolicy(common.NaNSkip)); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("all NaN: err = %v, want ErrEmpty", err)
	}
}
func TestMinMaxNaNError(t *testing.T) {
	if _, err := Max([]float64{1, math.NaN()}, common.WithNaNPolicy(common.NaNError)); !errors.Is(err, common.ErrNaN) {
		t.Fatalf("err = %v, want ErrNaN", err)
	}
}
func TestMinMaxRejectsBadOptions(t *testing.T) {
	if _, err := Min([]int{1}, common.WithPrecision(common.MaxPrecision+1)); !errors.Is(err, common.ErrPrecision) {
		t.Fatalf("err = %v, want ErrPrecision", err)
	}
}
