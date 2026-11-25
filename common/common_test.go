package common

import (
	"math"
	"math/big"
	"testing"
)

func TestIsFloat(t *testing.T) {
	if !IsFloat[float32]() || !IsFloat[float64]() {
		t.Fatal("float types not detected")
	}
	if IsFloat[int]() || IsFloat[int64]() || IsFloat[uint8]() {
		t.Fatal("integer type reported as float")
	}
}
func TestIsUnsigned(t *testing.T) {
	for _, got := range []bool{
		IsUnsigned[uint8](), IsUnsigned[uint16](), IsUnsigned[uint32](),
		IsUnsigned[uint64](), IsUnsigned[uintptr](), IsUnsigned[uint](),
	} {
		if !got {
			t.Fatal("unsigned type reported as signed")
		}
	}
	for _, got := range []bool{
		IsUnsigned[int8](), IsUnsigned[int16](), IsUnsigned[int32](),
		IsUnsigned[int64](), IsUnsigned[int](), IsUnsigned[float64](),
	} {
		if got {
			t.Fatal("signed or float type reported as unsigned")
		}
	}
}
func TestRoundTo(t *testing.T) {
	tests := []struct {
		name   string
		in     float64
		digits int
		want   float64
	}{
		{"half away from zero", 2.5, 0, 3},
		{"negative half away from zero", -2.5, 0, -3},
		{"two digits down", 1.2349, 2, 1.23},
		{"three digits up", 1.2349, 3, 1.235},
		{"carries into integer part", 1.9999, 2, 2},
		{"negative digits disable rounding", 1.2349, -1, 1.2349},
		{"magnitude too large to scale", 1e300, 10, 1e300},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := RoundTo(tc.in, tc.digits); got != tc.want {
				t.Fatalf("RoundTo(%v, %d) = %v, want %v", tc.in, tc.digits, got, tc.want)
			}
		})
	}
	if got := RoundTo(math.NaN(), 2); !math.IsNaN(got) {
		t.Fatalf("RoundTo(NaN) = %v, want NaN", got)
	}
	if got := RoundTo(math.Inf(-1), 2); !math.IsInf(got, -1) {
		t.Fatalf("RoundTo(-Inf) = %v, want -Inf", got)
	}
}
func TestFits(t *testing.T) {
	if v, ok := Fits[int8](bigFromString(t, "127")); !ok || v != 127 {
		t.Fatalf("127 into int8 = %v, %v", v, ok)
	}
	if _, ok := Fits[int8](bigFromString(t, "128")); ok {
		t.Fatal("128 reported as fitting int8")
	}
	if v, ok := Fits[uint8](bigFromString(t, "255")); !ok || v != 255 {
		t.Fatalf("255 into uint8 = %v, %v", v, ok)
	}
	if _, ok := Fits[uint8](bigFromString(t, "-1")); ok {
		t.Fatal("-1 reported as fitting uint8")
	}
	if !fitsMust[uint64](t, "18446744073709551615") {
		t.Fatal("MaxUint64 reported as not fitting uint64")
	}
	if _, ok := Fits[int64](bigFromString(t, "9223372036854775808")); ok {
		t.Fatal("2^63 reported as fitting int64")
	}
}

func bigFromString(t *testing.T, s string) *big.Int {
	t.Helper()
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("bad decimal %q", s)
	}
	return x
}
func fitsMust[T Number](t *testing.T, s string) bool {
	t.Helper()
	_, ok := Fits[T](bigFromString(t, s))
	return ok
}
