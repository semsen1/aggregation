package numbers

import (
	"errors"
	"github.com/semsen1/aggregation-go/common"
	"math"
	"math/rand"
	"testing"
)

func TestSumEmpty(t *testing.T) {
	if got, err := Sum([]int(nil)); err != nil || got != 0 {
		t.Fatalf("Sum(nil) = %v, %v; want 0, nil", got, err)
	}
	if got, err := Sum([]float64{}); err != nil || got != 0 {
		t.Fatalf("Sum(empty) = %v, %v; want 0, nil", got, err)
	}
}
func TestSumIntegers(t *testing.T) {
	if got, err := Sum([]int{1, 2, 3}); err != nil || got != 6 {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, err := Sum([]int64{1, -2, 3}); err != nil || got != 2 {
		t.Fatalf("got %v, %v", got, err)
	}
	if got, err := Sum([]uint8{1, 2, 3}); err != nil || got != 6 {
		t.Fatalf("got %v, %v", got, err)
	}
}
func TestSumIntOverflowNotSilent(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{"int8", func() error { _, err := Sum([]int8{100, 100}); return err }},
		{"uint8", func() error { _, err := Sum([]uint8{200, 100}); return err }},
		{"int64", func() error { _, err := Sum([]int64{common.MaxInt64, 1}); return err }},
		{"uint64", func() error { _, err := Sum([]uint64{^uint64(0), 1}); return err }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.run(); !errors.Is(err, common.ErrOverflow) {
				t.Fatalf("err = %v, want ErrOverflow", err)
			}
		})
	}
}
func TestSumIntExactBeyondInt64(t *testing.T) {
	// The running total leaves int64 twice and comes back; the result must be exact.
	got, err := Sum([]int64{common.MaxInt64, common.MaxInt64, -common.MaxInt64, -common.MaxInt64, 7})
	if err != nil || got != 7 {
		t.Fatalf("got %v, %v; want 7", got, err)
	}
}
func TestSumUint64BigPath(t *testing.T) {
	s := []uint64{^uint64(0), 1, 5}
	if _, err := Sum(s); !errors.Is(err, common.ErrOverflow) {
		t.Fatalf("err = %v, want ErrOverflow", err)
	}
	if got := SumBig(s).String(); got != "18446744073709551621" {
		t.Fatalf("SumBig = %s", got)
	}
}
func TestSumIntMatchesBigReference(t *testing.T) {
	rnd := rand.New(rand.NewSource(1))
	for iter := 0; iter < 300; iter++ {
		s := make([]int8, rnd.Intn(20))
		for i := range s {
			s[i] = int8(rnd.Intn(256) - 128)
		}
		want := SumBig(s)
		got, err := Sum(s)
		if want.IsInt64() && want.Int64() >= -128 && want.Int64() <= 127 {
			if err != nil {
				t.Fatalf("Sum(%v): unexpected %v", s, err)
			}
			if int64(got) != want.Int64() {
				t.Fatalf("Sum(%v) = %d, want %d", s, got, want.Int64())
			}
			continue
		}
		if !errors.Is(err, common.ErrOverflow) {
			t.Fatalf("Sum(%v) = %v, %v; want ErrOverflow (exact %v)", s, got, err, want)
		}
	}
}
func TestSumFloatCompensated(t *testing.T) {
	s := []float64{1, 1e100, 1, -1e100}
	naive := 0.0
	for _, v := range s {
		naive += v
	}
	if naive != 0 {
		t.Fatalf("sanity: naive sum = %v, want 0", naive)
	}
	got, err := Sum(s)
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if got != 2 {
		t.Fatalf("Sum = %v, want 2", got)
	}
}
func TestSumFloatOverflow(t *testing.T) {
	if _, err := Sum([]float64{math.MaxFloat64, math.MaxFloat64}); !errors.Is(err, common.ErrOverflow) {
		t.Fatalf("float64 overflow not reported")
	}
	if _, err := Sum([]float32{3e38, 3e38}); !errors.Is(err, common.ErrOverflow) {
		t.Fatalf("float32 overflow not reported")
	}
}
func TestSumInfPropagates(t *testing.T) {
	got, err := Sum([]float64{1, math.Inf(1)})
	if err != nil || !math.IsInf(got, 1) {
		t.Fatalf("got %v, %v; want +Inf, nil", got, err)
	}
}
func TestSumNaNPolicy(t *testing.T) {
	nan := math.NaN()
	if got, err := Sum([]float64{1, nan, 2}); err != nil || !math.IsNaN(got) {
		t.Fatalf("propagate: got %v, %v", got, err)
	}
	got, err := Sum([]float64{1, nan, 2}, common.WithNaNPolicy(common.NaNSkip))
	if err != nil || got != 3 {
		t.Fatalf("skip: got %v, %v; want 3", got, err)
	}
	if _, err := Sum([]float64{1, nan}, common.WithNaNPolicy(common.NaNError)); !errors.Is(err, common.ErrNaN) {
		t.Fatalf("error policy: got %v", err)
	}
}
func TestSumAlloc(t *testing.T) {
	s := make([]int64, 1024)
	for i := range s {
		s[i] = int64(i)
	}
	if n := testing.AllocsPerRun(100, func() { _, _ = Sum(s) }); n != 0 {
		t.Fatalf("Sum allocates %v times per call", n)
	}
}
