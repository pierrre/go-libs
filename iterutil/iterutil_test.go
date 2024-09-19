package iterutil

import (
	"fmt"
	"slices"
	"testing"

	"github.com/pierrre/assert"
)

func ExampleSeqToSeq2() {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		fmt.Println(k, v)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
}

func ExampleSeq2ToSeq() {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		fmt.Println(kv.Key, kv.Val)
	}
	// Output:
	// 0 zero
	// 1 one
	// 2 two
}

func TestSeqToSeq2(t *testing.T) {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		assert.Equal(t, v, kvs[k].Val)
	}
}

func TestSeqToSeq2Stop(t *testing.T) {
	kvs := []KeyVal[int, string]{
		NewKeyVal(0, "zero"),
		NewKeyVal(1, "one"),
		NewKeyVal(2, "two"),
	}
	it := SeqToSeq2(slices.Values(kvs), KeyVal[int, string].Values)
	for k, v := range it {
		assert.Equal(t, v, kvs[k].Val)
		break
	}
}

func TestSeq2ToSeq(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		assert.Equal(t, kv.Val, ss[kv.Key])
	}
}

func TestSeq2ToSeqStop(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeq(slices.All(ss), NewKeyVal)
	for kv := range it {
		assert.Equal(t, kv.Val, ss[kv.Key])
		break
	}
}
