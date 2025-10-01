package simulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHello(t *testing.T) {
	t.Parallel()

	got := Hello()
	want := "Hello, World!"
	assert.Equal(t, want, got)
}
