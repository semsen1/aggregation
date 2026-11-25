package numbers

import (
	"github.com/semsen1/aggregation-go/common"
	"math"
	"math/big"
)

func sum[T common.Number](s []T, c common.Config) (T, error) {
	var zero T
	if common.IsFloat[T]() {
		a, err := common.AccumulateFloat(s, nil, c.NaN)
		if err != nil {
			return zero, err
		}
		total := a.Total()
		if math.IsInf(total, 0) && !a.SawInf {
			return zero, common.ErrOverflow
		}
		if a.Overflow && a.Special == 0 {
			return zero, common.ErrOverflow
		}
		res := T(total)
		if !math.IsInf(total, 0) && math.IsInf(float64(res), 0) {
			return zero, common.ErrOverflow
		}

		return res, nil
	}
	acc, bigAcc, _ := common.AccumulateInt(s, nil)
	if bigAcc != nil {
		return common.Narrow[T](bigAcc)
	}

	return common.NarrowInt64[T](acc)
}

// Sum returns the sum of s.
//
// Integer input is summed exactly. The running total is kept in int64 and is
// promoted to math/big as soon as it would overflow, so Sum never wraps. If the
// exact total is still not representable in T — Sum([]int8{100, 100} or
// Sum([]int64{math.MaxInt64, 1}) — the result is ErrOverflow. Use SumBig when
// the exact value is wanted regardless of type.
//
// Float input is summed in float64 with Kahan-Babuška-Neumaier compensation,
// then converted to T. A float32 slice therefore gains precision from the wider
// accumulator, and the call reports ErrOverflow if the converted total leaves
// the float32 range or if a finite input summed past float64's range. An Inf
// present in the input propagates as Inf without an error, and NaN follows the
// NaNPolicy option.
//
// An empty slice sums to zero, not an error: the sum over no elements is 0.
// Allocations are zero unless an integer total leaves int64. The
// WithPrecision option is accepted for symmetry with Avg but has no effect;
// an out-of-range value is still reported as ErrPrecision.
func Sum[T common.Number](s []T, opts ...common.Option) (T, error) {
	if len(opts) == 0 {
		return sum(s, common.DefaultConfig)
	}
	c, err := common.BuildConfig(opts)
	if err != nil {
		var zero T
		return zero, err
	}
	return sum(s, c)
}

// SumBig returns the exact sum of s as an arbitrary-precision integer. It is
// the escape hatch for totals that do not fit any Go integer type: for
// Sum([]uint64{math.MaxUint64, 1}, which reports ErrOverflow, SumBig returns
// 2^64. It never fails and never wraps.
func SumBig[T common.Integer](s []T) *big.Int {
	acc, bigAcc, _ := common.AccumulateInt[T](s, nil)
	if bigAcc != nil {
		return bigAcc
	}
	return big.NewInt(acc)
}
