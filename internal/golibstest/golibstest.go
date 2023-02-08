// Package golibstest provides internal tools for tests.
package golibstest

import (
	"github.com/pierrre/assert/ext/davecghspew"
	"github.com/pierrre/assert/ext/pierrrecompare"
	"github.com/pierrre/assert/ext/pierrreerrors"
)

// Configure configures tools used in tests.
func Configure() {
	pierrrecompare.Configure()
	davecghspew.ConfigureDefault()
	pierrreerrors.Configure()
}
