package numbers

import (
	"github.com/semsen1/aggregation-go/common"
	"math"
	"math/big"
)

// Avg returns the arithmetic mean of s as a float64.
//
// The result type is always float64, never T: an average of ints is generally
// not an int, and returning T would make Avg([]int{3, 4}) equal 3. For integer
// input the mean is derived from the exact sum, so it does not drift when the
// running total passes 2^53.
//
// Float input is summed with Kahan-Babuška-Neumaier compensation, so tiny
// values are not swallowed by large ones: the mean of {1, 1e100, 1, -1e100} is
// 0.5, where a naive loop returns 0.
//
// WithPrecision rounds the result to a fixed number of decimal digits; without
// it the full float64 value is returned. The result is ErrEmpty for an empty
// slice, or when the NaN policy filtered every element out, and ErrOverflow if
// float input summed past the float64 range without an Inf in the input.
func Avg[T common.Number](s []T, opts ...common.Option) (float64, error) {
	if len(opts) == 0 {
		return avgOf(s, nil, common.DefaultConfig)
	}
	c, err := common.BuildConfig(opts)
	if err != nil {
		return 0, err
	}
	return avgOf(s, nil, c)
}

// AvgIf returns the mean of the elements of s for which keep reports true.
//
// It is a single pass and does not allocate a filtered slice. The predicate is
// applied before the NaN policy, so it sees NaN values as they are. A nil
// predicate yields ErrNilFunc; a predicate that selects nothing yields
// ErrEmpty, exactly as an empty slice would.
func AvgIf[T common.Number](s []T, keep func(T) bool, opts ...common.Option) (float64, error) {
	if keep == nil {
		return 0, common.ErrNilFunc
	}
	if len(opts) == 0 {
		return avgOf(s, keep, common.DefaultConfig)
	}
	c, err := common.BuildConfig(opts)
	if err != nil {
		return 0, err
	}
	return avgOf(s, keep, c)
}
func avgOf[T common.Number](s []T, keep func(T) bool, c common.Config) (float64, error) {
	if common.IsFloat[T]() {
		a, err := common.AccumulateFloat(s, keep, c.NaN)
		if err != nil {
			return 0, err
		}
		if a.N == 0 {
			return 0, common.ErrEmpty
		}
		total := a.Total()
		if math.IsInf(total, 0) && !a.SawInf {
			return 0, common.ErrOverflow
		}
		return common.RoundTo(total/float64(a.N), c.Digits), nil
	}
	acc, bigAcc, n := common.AccumulateInt(s, keep)
	if n == 0 {
		return 0, common.ErrEmpty
	}
	return common.RoundTo(common.IntegerAvg(acc, bigAcc, n), c.Digits), nil
}

var _ = big.NewInt
