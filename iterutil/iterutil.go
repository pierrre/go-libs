// Package iterutil provides utilities for iterators.
package iterutil

import (
	"iter"
)

// SeqToSeq2 converts a [iter.Seq] to a [iter.Seq2].
func SeqToSeq2[In, Out1, Out2 any](in iter.Seq[In], convert func(In) (Out1, Out2)) iter.Seq2[Out1, Out2] {
	return func(yield func(Out1, Out2) bool) {
		in(func(i In) bool {
			return yield(convert(i))
		})
	}
}

// SeqToSeq2Index converts a [iter.Seq] to a [iter.Seq2] that yields the index and value.
// The index is the position of the value in the sequence, starting from 0.
func SeqToSeq2Index[V any](in iter.Seq[V]) iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		i := 0
		in(func(v V) bool {
			ok := yield(i, v)
			i++
			return ok
		})
	}
}

// Seq2ToSeq converts a [iter.Seq2] to a [iter.Seq].
func Seq2ToSeq[In1, In2, Out any](in iter.Seq2[In1, In2], convert func(In1, In2) Out) iter.Seq[Out] {
	return func(yield func(Out) bool) {
		in(func(i1 In1, i2 In2) bool {
			return yield(convert(i1, i2))
		})
	}
}

// Seq2ToSeqKey converts a [iter.Seq2] to a [iter.Seq] that yields only the keys.
func Seq2ToSeqKey[K, V any](seq iter.Seq2[K, V]) iter.Seq[K] {
	return Seq2ToSeq(seq, func(k K, _ V) K {
		return k
	})
}

// Seq2ToSeqValue converts a [iter.Seq2] to a [iter.Seq] that yields only the values.
func Seq2ToSeqValue[K, V any](seq iter.Seq2[K, V]) iter.Seq[V] {
	return Seq2ToSeq(seq, func(_ K, v V) V {
		return v
	})
}

// KeyVal represents a key-value pair.
type KeyVal[K, V any] struct {
	Key K
	Val V
}

// NewKeyVal creates a new [KeyVal].
func NewKeyVal[K, V any](key K, val V) KeyVal[K, V] {
	return KeyVal[K, V]{
		Key: key,
		Val: val,
	}
}

// Values returns the key and value of the [KeyVal].
func (kv KeyVal[K, V]) Values() (key K, val V) {
	return kv.Key, kv.Val
}

// Repeat returns a new [iter.Seq] that repeats the input sequence n times.
// If n is negative, the sequence is repeated indefinitely.
func Repeat[V any](seq iter.Seq[V], n int) iter.Seq[V] {
	return func(yield func(V) bool) {
		for i, ok := 0, true; ok && (n < 0 || i < n); i++ {
			ok = false
			seq(func(v V) bool {
				ok = yield(v)
				return ok
			})
		}
	}
}

// Repeat2 returns a new [iter.Seq2] that repeats the input sequence n times.
// If n is negative, the sequence is repeated indefinitely.
func Repeat2[K, V any](seq iter.Seq2[K, V], n int) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i, ok := 0, true; ok && (n < 0 || i < n); i++ {
			ok = false
			seq(func(k K, v V) bool {
				ok = yield(k, v)
				return ok
			})
		}
	}
}
