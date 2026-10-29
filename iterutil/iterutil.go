// Package iterutil provides utilities for iterators.
package iterutil

import (
	"iter"
	"sync/atomic"
	"time"
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

// WithConsumerStallTimeout returns a new [iter.Seq] that stops if the consumer holds a value for longer than stallTimeout.
// Time spent by seq between two values is not watched: only the time the consumer holds a value counts.
// If the consumer holds a value for longer than stallTimeout, onStall is called once, on the timer goroutine, and no further values are delivered.
// Use onStall to release resources pinned by a stalled consumer, for example by canceling the context behind an open database cursor.
// The returned sequence can be ranged multiple times, but ranges must not overlap: a stall cannot be attributed to a range.
// To consume a sequence with multiple consumers, wrap it once per consumer.
// onStall is called on the timer goroutine while the consumer's loop is parked inside the body, so it must be safe to call concurrently with the range.
// For key-value sequences, see [WithConsumerStallTimeout2].
// onStall must not be nil, and stallTimeout must be positive.
func WithConsumerStallTimeout[V any](seq iter.Seq[V], stallTimeout time.Duration, onStall func()) iter.Seq[V] {
	if onStall == nil {
		panic("iterutil: onStall must not be nil")
	}
	if stallTimeout <= 0 {
		panic("iterutil: stallTimeout must be positive")
	}
	return func(yield func(V) bool) {
		var stalled atomic.Bool
		watchdog := time.AfterFunc(stallTimeout, func() {
			if stalled.CompareAndSwap(false, true) {
				onStall()
			}
		})
		defer watchdog.Stop()
		watchdog.Stop()
		seq(func(v V) bool {
			if stalled.Load() {
				return false
			}
			watchdog.Reset(stallTimeout)
			ok := yield(v)
			watchdog.Stop()
			return ok
		})
	}
}

// WithConsumerStallTimeout2 is [WithConsumerStallTimeout] for key-value sequences.
// See [WithConsumerStallTimeout] for the stall semantics.
func WithConsumerStallTimeout2[K, V any](seq iter.Seq2[K, V], stallTimeout time.Duration, onStall func()) iter.Seq2[K, V] {
	return SeqToSeq2(
		WithConsumerStallTimeout(Seq2ToSeq(seq, NewKeyVal[K, V]), stallTimeout, onStall),
		KeyVal[K, V].Values,
	)
}
