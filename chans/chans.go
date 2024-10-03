package chans

import (
	"iter"
)

func Values[C ~<-chan E, E any](ch C) iter.Seq[E] {
	return func(yield func(E) bool) {
		for v := range ch {
			if !yield(v) {
				break
			}
		}
	}
}

func SendSeq[C ~chan<- E, E any](ch C, seq iter.Seq[E]) {
	for v := range seq {
		ch <- v
	}
}
