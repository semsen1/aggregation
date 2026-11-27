package numbers

import (
	"github.com/semsen1/aggregation-go/common"
	"math"
)

// Min returns the smallest element of s.
//
// An empty slice yields ErrEmpty: there is no smallest element of nothing, and
// returning 0 would quietly poison downstream comparisons.
//
// NaN handling follows NaNPolicy and is independent of NaN's position. A naive
// loop returns whatever the comparison happens to do — Min([]float64{NaN, 1})
// gives NaN while Min([]float64{1, NaN}) gives 1 — which makes results depend
// on input order. Here the default NaNPropagate returns NaN for either input,
// NaNSkip ignores NaN values, and NaNError returns ErrNaN.
func Min[T common.Number](s []T, opts ...common.Option) (T, error) {
	return extremum(s, true, opts)
}

// Max returns the largest element of s, with the same contract as Min.
func Max[T common.Number](s []T, opts ...common.Option) (T, error) {
	return extremum(s, false, opts)
}
func extremum[T common.Number](s []T, wantMin bool, opts []common.Option) (T, error) {
	var zero T
	c := common.DefaultConfig
	if len(opts) > 0 {
		var err error
		c, err = common.BuildConfig(opts)
		if err != nil {
			return zero, err
		}
	}
	floating := common.IsFloat[T]()
	var best T
	found := false
	sawNaN := false
	for _, v := range s {
		if floating && math.IsNaN(float64(v)) {
			switch c.NaN {
			case common.NaNSkip:
				continue
			case common.NaNError:
				return zero, common.ErrNaN
			default:
				sawNaN = true
				continue
			}
		}
		if !found {
			best, found = v, true
			continue
		}
		if wantMin {
			if v < best {
				best = v
			}
		} else if v > best {
			best = v
		}
	}
	if sawNaN {
		return T(math.NaN()), nil
	}
	if !found {
		return zero, common.ErrEmpty
	}
	return best, nil
}
