package common

import (
	"errors"
	"fmt"
	"math"
	"math/big"
)

// Integer is the set of Go built in integer types. The tilde form admits
// named types with those underlying types, so []time.Duration and
// []type Money int64 both work.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Float is the set of Go built in floating-point types.
type Float interface {
	~float32 | ~float64
}

// Number is any numeric type this package can aggregate.
type Number interface {
	Integer | Float
}

// MaxPrecision is the largest number of decimal digits after the point that
// WithPrecision accepts. It matches the decimal range of float64.
const MaxPrecision = 308

// Errors returned by this package. Sentinels are always wrapped with
// fmt.Errorf, so test them with errors.Is rather than ==.
var (
	// ErrEmpty reports that an aggregate is undefined for the given input,
	// typically because the slice (or the subset selected by a predicate)
	// contained no elements.
	ErrEmpty = errors.New("aggregation: empty input")

	// ErrOverflow reports that the exact result is not representable in the
	// element type. It is returned instead of a silently wrapped or infinite
	// value.
	ErrOverflow = errors.New("aggregation: exact result is not representable in the element type")

	// ErrPrecision reports a precision outside [0, MaxPrecision].
	ErrPrecision = errors.New("aggregation: precision out of range [0, 308]")

	// ErrNaN reports a NaN in the input when NaNError policy is selected.
	ErrNaN = errors.New("aggregation: NaN in input")

	// ErrNilFunc reports a nil predicate passed to a *If function.
	ErrNilFunc = errors.New("aggregation: nil predicate")

	// ErrNaNPolicy reports an unknown NaNPolicy value.
	ErrNaNPolicy = errors.New("aggregation: unknown NaN policy")
)

// NaNPolicy decides how floating-point NaN values in the input are treated.
// It has no effect on integer input, which cannot contain NaN.
type NaNPolicy int

const (
	// NaNPropagate is the default and follows IEEE 754: a NaN anywhere in the
	// input makes the result NaN. Unlike a naive loop, the outcome does not
	// depend on where the NaN sits, which matters for Min and Max.
	NaNPropagate NaNPolicy = iota

	// NaNSkip ignores NaN values. They are not counted, so an input that is
	// entirely NaN behaves as empty input.
	NaNSkip

	// NaNError aborts the call with ErrNaN as soon as a NaN is seen.
	NaNError
)

// Config is the resolved option set for a single call.
type Config struct {
	Digits int // -1 means "do not round"
	NaN    NaNPolicy
	Err    error
}

var DefaultConfig = Config{Digits: -1, NaN: NaNPropagate}

// Option tunes a single aggregation call. Nil options are ignored.
type Option func(*Config)

// WithPrecision rounds the result of Avg and AvgIf to digits decimal places.
// Rounding is half away from zero, applied to the binary value, so the usual
// caveat holds: 1.005 is really 1.00499999999999989 in float64 and rounds to
// 1.00, not 1.01.
//
// A negative value disables rounding and returns the full float64 result,
// which is also the default. Values above MaxPrecision yield ErrPrecision.
func WithPrecision(digits int) Option {
	return func(c *Config) {
		if c.Err != nil {
			return
		}

		if digits < -1 || digits > MaxPrecision {
			c.Err = fmt.Errorf("avg: precision %d out of range [-1, %d]: %w",
				digits, MaxPrecision, ErrPrecision)
			return
		}
		c.Digits = digits
	}
}

// WithNaNPolicy selects how NaN values in floating-point input are handled.
// The default is NaNPropagate.
func WithNaNPolicy(p NaNPolicy) Option {
	return func(c *Config) { c.NaN = p }
}

// BuildConfig resolves options. It is only called when at least one option is
// present, which keeps the no-option path free of allocations.
func BuildConfig(opts []Option) (Config, error) {
	c := DefaultConfig
	for _, o := range opts {
		if o != nil {
			o(&c)
		}
	}
	if c.Err != nil {
		return Config{}, c.Err
	}
	if c.Digits < -1 || c.Digits > MaxPrecision {
		return Config{}, fmt.Errorf("%w: %d", ErrPrecision, c.Digits)
	}
	if c.NaN < NaNPropagate || c.NaN > NaNError {
		return Config{}, fmt.Errorf("%w: %d", ErrNaNPolicy, c.NaN)
	}
	return c, nil
}

const (
	MaxInt64  int64  = 1<<63 - 1
	MinInt64  int64  = -1 << 63
	MaxInt64u uint64 = 1<<63 - 1
)

// IsUnsigned reports whether T is an unsigned integer type. It folds to a
// constant for a concrete T and returns false for float types.
func IsUnsigned[T Number]() bool {
	var zero T
	return zero-1 > zero
}

// IsFloat reports whether T is a floating-point type.
func IsFloat[T Number]() bool {
	var one T = 1
	return one/2 != 0
}

// toBig converts a single integer element to a big.Int without truncation.
// It must not be called with a float type.
func toBig[T Number](v T) *big.Int {
	if IsUnsigned[T]() {
		return new(big.Int).SetUint64(uint64(v))
	}
	return big.NewInt(int64(v))
}

// AccumulateInt sums the selected elements exactly. It keeps a fast int64
// accumulator and switches to math/big the first time the running total would
// overflow, so the returned total never wraps.
func AccumulateInt[T Number](s []T, keep func(T) bool) (acc int64, bigAcc *big.Int, n uint64) {
	unsigned := IsUnsigned[T]()
	for _, v := range s {
		if keep != nil && !keep(v) {
			continue
		}
		n++
		if bigAcc != nil {
			bigAcc.Add(bigAcc, toBig(v))
			continue
		}

		if unsigned && uint64(v) > MaxInt64u {
			bigAcc = new(big.Int).SetUint64(uint64(v))
			bigAcc.Add(bigAcc, new(big.Int).SetInt64(acc))
			continue
		}
		x := int64(v)
		if (x > 0 && acc > MaxInt64-x) || (x < 0 && acc < MinInt64-x) {
			bigAcc = new(big.Int).SetInt64(acc)
			bigAcc.Add(bigAcc, big.NewInt(x))
			continue
		}
		acc += x
	}
	return acc, bigAcc, n
}

// FloatAccum is a compensated accumulator for float64 values.
type FloatAccum struct {
	Sum, Comp float64
	Special   float64
	N         uint64
	SawInf    bool
	Overflow  bool
}

func (a *FloatAccum) Add(x float64) {
	a.N++
	if math.IsNaN(x) || math.IsInf(x, 0) {
		a.Special += x
		a.SawInf = a.SawInf || !math.IsNaN(x)
		return
	}
	t := a.Sum + x
	if math.IsInf(t, 0) {
		a.Overflow = true
		return
	}
	if math.Abs(a.Sum) >= math.Abs(x) {
		a.Comp += (a.Sum - t) + x
	} else {
		a.Comp += (x - t) + a.Sum
	}
	a.Sum = t
}

func (a FloatAccum) Total() float64 {
	switch {
	case a.Special != 0:
		return a.Special
	case a.Overflow:
		return math.Inf(1)
	}
	return a.Sum + a.Comp
}

func (a FloatAccum) Result() (float64, error) {
	if a.Special != 0 {
		return a.Special, nil
	}
	if a.Overflow {
		return 0, ErrOverflow
	}
	res := a.Sum + a.Comp
	if math.IsInf(res, 0) {
		return 0, ErrOverflow
	}
	return res, nil
}

// AccumulateFloat sums the selected elements using Kahan-Babuška-Neumaier
// compensation. A predicate, if given, is applied before the NaN policy, so it
// sees the raw value.
func AccumulateFloat[T Number](s []T, keep func(T) bool, nan NaNPolicy) (FloatAccum, error) {
	var a FloatAccum
	for _, v := range s {
		if keep != nil && !keep(v) {
			continue
		}
		x := float64(v)
		if math.IsNaN(x) {
			switch nan {
			case NaNSkip:
				continue
			case NaNError:
				return a, ErrNaN
			}
		}
		a.Add(x)
	}
	return a, nil
}

// Fits reports whether x is representable in T and returns the converted value.
// It works by narrow-conversion round trip, which needs no per-type bounds table.
// X must be non-nil
func Fits[T Number](x *big.Int) (T, bool) {
	var zero T
	unsigned := IsUnsigned[T]()
	if x.IsInt64() {
		v := x.Int64()
		if unsigned && v < 0 {
			return zero, false
		}
		t := T(v)
		if int64(t) == v {
			return t, true
		}
		return zero, false
	}
	if !unsigned || !x.IsUint64() {
		return zero, false
	}
	u := x.Uint64()
	t := T(u)
	if uint64(t) == u {
		return t, true
	}
	return zero, false
}

// Narrow converts an exact integer total to T or reports ErrOverflow.
func Narrow[T Number](x *big.Int) (T, error) {
	var zero T
	if t, ok := Fits[T](x); ok {
		return t, nil
	}
	return zero, fmt.Errorf("%w: %s", ErrOverflow, x)
}

// NarrowInt64 converts an int64 total to T or reports ErrOverflow. This is the
// hot path: an exact total that still Fits int64.
func NarrowInt64[T Number](v int64) (T, error) {
	var zero T
	if IsUnsigned[T]() && v < 0 {
		return zero, ErrOverflow
	}
	t := T(v)
	if int64(t) != v {
		return zero, ErrOverflow
	}
	return t, nil
}

// RoundTo rounds x to digits decimal places, half away from zero. A negative
// digits value, NaN and Inf are returned unchanged, as is a magnitude too large
// for the scaling to stay finite.
func RoundTo(x float64, digits int) float64 {
	if digits < 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	if digits == 0 {
		return math.Round(x)
	}
	p := math.Pow10(digits)
	scaled := x * p
	if math.IsInf(scaled, 0) {
		return x
	}
	return math.Round(scaled) / p
}

// IntegerAvg divides an exact integer total by n using big.Rat when the total
// has left int64, so the mean is correctly rounded.
func IntegerAvg(acc int64, bigAcc *big.Int, n uint64) float64 {
	if bigAcc == nil {
		return float64(acc) / float64(n)
	}
	r := new(big.Rat).SetFrac(bigAcc, new(big.Int).SetUint64(n))
	f, _ := r.Float64()
	return f
}
