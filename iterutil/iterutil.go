package iterutil

import (
	"iter"
)

func SeqToSeq2[In, Out1, Out2 any](in iter.Seq[In], convert func(In) (Out1, Out2)) iter.Seq2[Out1, Out2] {
	return func(yield func(Out1, Out2) bool) {
		for i := range in {
			o1, o2 := convert(i)
			if !yield(o1, o2) {
				break
			}
		}
	}
}

func Seq2ToSeq[In1, In2, Out any](in iter.Seq2[In1, In2], convert func(In1, In2) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		for i1, i2 := range in {
			if !yield(convert(i1, i2)) {
				break
			}
		}
	}
}

type KeyVal[K, V any] struct {
	Key K
	Val V
}

func NewKeyVal[K, V any](key K, val V) KeyVal[K, V] {
	return KeyVal[K, V]{
		Key: key,
		Val: val,
	}
}

func (kv KeyVal[K, V]) Values() (key K, val V) {
	return kv.Key, kv.Val
}
