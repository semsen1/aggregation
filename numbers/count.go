package numbers

import "github.com/semsen1/aggregation-go/common"

// CountIf returns how many elements of s satisfy keep.
//
// The count is uint64 because a length is bounded by int on a 64-bit platform
// but not on a 32-bit one, and a counter that silently overflows is exactly the
// class of bug this package exists to avoid.
//
// An empty slice counting zero is not an error. A nil predicate is ErrNilFunc.
// CountIf accepts any element type, not just numbers, and performs one pass
// with no allocations.
func CountIf[T any](s []T, keep func(T) bool) (uint64, error) {
	if keep == nil {
		return 0, common.ErrNilFunc
	}
	var n uint64
	for _, v := range s {
		if keep(v) {
			n++
		}
	}
	return n, nil
}
