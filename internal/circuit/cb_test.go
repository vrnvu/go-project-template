package circuit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsserts(t *testing.T) {
	t.Parallel()
	asserts(true)

	assert.Panics(t, func() {
		asserts(false)
	})
}
