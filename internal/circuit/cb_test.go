package circuit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsserts(t *testing.T) {
	asserts(true)

	assert.Panics(t, func() {
		asserts(false)
	})
}
