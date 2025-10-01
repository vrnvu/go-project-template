package circuit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeCBInvalid(t *testing.T) {
	c, err := NewTimeCB(0*time.Second, 1, 1)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "openTimeout")

	c, err = NewTimeCB(1*time.Second, 0, 0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "halfOpenProbesThreshold")

	c, err = NewTimeCB(1*time.Second, 1, 0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "closedFailuresThreshold")
}
