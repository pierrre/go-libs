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

func TestSeqToSeq2Index(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := SeqToSeq2Index(slices.Values(ss))
	for k, v := range it {
		assert.Equal(t, v, ss[k])
	}
}

func TestSeqToSeq2IndexStop(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := SeqToSeq2Index(slices.Values(ss))
	for k, v := range it {
		assert.Equal(t, v, ss[k])
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

func TestSeq2ToSeqKey(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeqKey(slices.All(ss))
	expected := 0
	for k := range it {
		assert.Equal(t, k, expected)
		expected++
	}
	assert.Equal(t, expected, len(ss))
}

func TestSeq2ToSeqValue(t *testing.T) {
	ss := []string{"zero", "one", "two"}
	it := Seq2ToSeqValue(slices.All(ss))
	i := 0
	for v := range it {
		assert.Equal(t, v, ss[i])
		i++
	}
	assert.Equal(t, i, len(ss))
}
