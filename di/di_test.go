//nolint:forbidigo // Allow to use fmt.Println in this test.
package di

import (
	"fmt"
	"io"
	"testing"
)

func Example() {
	// New container.
	c := new(Container)
	// Set ServiceA.
	Set(c, "", func(c *Container) (*ServiceA, Close, error) {
		return &ServiceA{}, nil, nil
	})
	// Set ServiceB.
	somethingWrong := false
	Set(c, "", func(c *Container) (*ServiceB, Close, error) {
		// We know that ServiceA's builder doesn't return an error, so we ignore it.
		sa := Must(Get[*ServiceA](c, ""))
		if somethingWrong {
			return nil, nil, fmt.Errorf("error")
		}
		sb := &ServiceB{
			sa.DoA,
		}
		return sb, sb.close, nil
	})
	// Set ServiceC.
	Set(c, "", func(c *Container) (*ServiceC, Close, error) {
		sb, err := Get[*ServiceB](c, "")
		if err != nil {
			return nil, nil, err
		}
		sc := &ServiceC{
			sb.DoB,
		}
		// The ServiceC close function doesn't return an error, so we wrap it.
		cl := func() error {
			sc.close()
			return nil
		}
		return sc, cl, nil
	})
	// Get ServiceC and call it.
	sc, err := Get[*ServiceC](c, "")
	if err != nil {
		panic(err)
	}
	sc.DoC()
	// Close container.
	c.Close(func(err error) {
		panic(err)
	})
	// Output:
	// do A
	// do B
	// do C
	// close B
	// close C
}

type ServiceA struct{}

func (sa *ServiceA) DoA() {
	fmt.Println("do A")
}

type ServiceB struct {
	sa func()
}

func (sb *ServiceB) DoB() {
	sb.sa()
	fmt.Println("do B")
}

func (sb *ServiceB) close() error {
	fmt.Println("close B")
	return nil
}

type ServiceC struct {
	sb func()
}

func (sc *ServiceC) DoC() {
	sc.sb()
	fmt.Println("do C")
}

func (sc *ServiceC) close() {
	fmt.Println("close C")
}

var benchmarkGetServiceNameResult string

func BenchmarkGetServiceNameString(b *testing.B) {
	var s string
	for i := 0; i < b.N; i++ {
		s = getServiceName[string]()
	}
	benchmarkGetServiceNameResult = s
}

func BenchmarkGetServiceNameIOWriter(b *testing.B) {
	var s string
	for i := 0; i < b.N; i++ {
		s = getServiceName[io.Writer]()
	}
	benchmarkGetServiceNameResult = s
}
