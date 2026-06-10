package sliceutil

import "slices"

// Reversed は s を反転したコピーを返す。元のスライスは変更しない。
func Reversed[S ~[]E, E any](s S) S {
	r := slices.Clone(s)
	slices.Reverse(r)
	return r
}
