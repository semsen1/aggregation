package numbers

import (
	"errors"
	"github.com/semsen1/aggregation-go/common"
	"math"
	"testing"
)

func TestAvgOfIntIsFloat(t *testing.T) {
	got, err := Avg([]int{3, 4})
	if err != nil {
		t.Fatalf("Avg: %v", err)
	}
	if got != 3.5 {
		t.Fatalf("Avg([]int{3,4}) = %v, want 3.5", got)
	}
}
func TestAvgEmpty(t *testing.T) {
	if _, err := Avg([]int(nil)); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("int: err = %v, want ErrEmpty", err)
	}
	if _, err := Avg([]float64{}); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("float: err = %v, want ErrEmpty", err)
	}
}
func TestAvgPrecision(t *testing.T) {
	tests := []struct {
		in     []float64
		digits int
		want   float64
	}{
		{[]float64{1.23456}, 2, 1.23},
		{[]float64{1.23456}, 4, 1.2346},
		{[]float64{1.5}, 0, 2},
		{[]float64{2.5}, 0, 3},
		{[]float64{1.23456}, -1, 1.23456},
	}
	for _, tc := range tests {
		got, err := Avg(tc.in, common.WithPrecision(tc.digits))
		if err != nil {
			t.Fatalf("Avg(%v, %d): %v", tc.in, tc.digits, err)
		}
		if got != tc.want {
			t.Fatalf("Avg(%v, %d) = %v, want %v", tc.in, tc.digits, got, tc.want)
		}
	}
}
func TestAvgPrecisionOutOfRange(t *testing.T) {
	if _, err := Avg([]int{1}, common.WithPrecision(-2)); !errors.Is(err, common.ErrPrecision) {
		t.Fatalf("err = %v, want ErrPrecision", err)
	}
	if _, err := Avg([]int{1}, common.WithPrecision(common.MaxPrecision+1)); !errors.Is(err, common.ErrPrecision) {
		t.Fatalf("err = %v, want ErrPrecision", err)
	}
}

func TestAvgCompensated(t *testing.T) {
	s := []float64{1, 1e100, 1, -1e100}
	naive := 0.0
	for _, v := range s {
		naive += v
	}
	if naive != 0 {
		t.Fatalf("sanity: naive mean = %v, want 0", naive/float64(len(s)))
	}
	got, err := Avg(s)
	if err != nil {
		t.Fatalf("Avg: %v", err)
	}
	if got != 0.5 {
		t.Fatalf("Avg = %v, want 0.5", got)
	}
}
func TestAvgExactIntegers(t *testing.T) {
	got, err := Avg([]int64{common.MaxInt64, common.MaxInt64})
	if err != nil {
		t.Fatalf("Avg: %v", err)
	}
	if want := float64(common.MaxInt64); got != want {
		t.Fatalf("Avg = %v, want %v", got, want)
	}
	got, err = Avg([]int64{common.MinInt64, common.MinInt64})
	if err != nil {
		t.Fatalf("Avg: %v", err)
	}
	if want := float64(common.MinInt64); got != want {
		t.Fatalf("Avg = %v, want %v", got, want)
	}
}
func TestAvgIf(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	got, err := AvgIf(s, func(v int) bool { return v%2 == 0 })
	if err != nil {
		t.Fatalf("AvgIf: %v", err)
	}
	if got != 4 {
		t.Fatalf("AvgIf = %v, want 4", got)
	}
}
func TestAvgIfErrors(t *testing.T) {
	if _, err := AvgIf([]int{1, 3}, func(int) bool { return false }); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("err = %v, want ErrEmpty", err)
	}
	if _, err := AvgIf([]int{1}, nil); !errors.Is(err, common.ErrNilFunc) {
		t.Fatalf("err = %v, want ErrNilFunc", err)
	}
}
func TestAvgNaNPolicy(t *testing.T) {
	nan := math.NaN()
	if got, err := Avg([]float64{1, nan}); err != nil || !math.IsNaN(got) {
		t.Fatalf("propagate: got %v, %v", got, err)
	}
	got, err := Avg([]float64{1, nan, 3}, common.WithNaNPolicy(common.NaNSkip))
	if err != nil || got != 2 {
		t.Fatalf("skip: got %v, %v; want 2", got, err)
	}
	if _, err := Avg([]float64{nan}, common.WithNaNPolicy(common.NaNSkip)); !errors.Is(err, common.ErrEmpty) {
		t.Fatalf("all NaN skipped: err = %v, want ErrEmpty", err)
	}
	if _, err := Avg([]float64{nan}, common.WithNaNPolicy(common.NaNError)); !errors.Is(err, common.ErrNaN) {
		t.Fatalf("error policy: err = %v, want ErrNaN", err)
	}
}
func TestAvgAlloc(t *testing.T) {
	s := make([]int64, 1024)
	for i := range s {
		s[i] = int64(i)
	}
	if n := testing.AllocsPerRun(100, func() { _, _ = Avg(s) }); n != 0 {
		t.Fatalf("Avg allocates %v times per call", n)
	}
}

func TestAvgFastPathMatchesSlow(t *testing.T) {
	s := []int64{1, 2, 3, 4}
	fast, err1 := Avg(s)
	slow, err2 := Avg(s, func(*common.Config) {})
	if err1 != nil || err2 != nil || fast != slow {
		t.Fatalf("fast=%v/%v slow=%v/%v", fast, err1, slow, err2)
	}
}
