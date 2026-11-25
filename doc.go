// Package aggregation aggregates numeric slices without the usual silent mistakes.
//
// Rules hold for every function in this package.
//
// Exactness. Integer sums are computed exactly: a running total that would
// overflow int64 is promoted to math/big, and if the exact result is still not
// representable in the element type the call fails with ErrOverflow instead of
// wrapping. Float sums use compensated (Kahan-Babuška-Neumaier) summation, so
// adding a small value to a huge one does not lose it.

package aggregation
