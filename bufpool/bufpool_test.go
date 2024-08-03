package bufpool

import (
	"bytes"
	"testing"
)

func Test(t *testing.T) {
	p := &Pool{}
	for range 10 {
		buf := p.Get()
		p.Put(buf)
	}
}

const testBenchData = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum." //nolint:lll // This is a long text for benchmark.

var benchmarkResult string

func BenchmarkBufPool(b *testing.B) {
	p := &Pool{}
	var s string
	for range b.N {
		buf := p.Get()
		for range 10 {
			buf.WriteString(testBenchData)
		}
		s = buf.String()
		p.Put(buf)
	}
	benchmarkResult = s
}

func BenchmarkBufWithoutPool(b *testing.B) {
	var s string
	for range b.N {
		buf := new(bytes.Buffer)
		for range 10 {
			buf.WriteString(testBenchData)
		}
		s = buf.String()
	}
	benchmarkResult = s
}
