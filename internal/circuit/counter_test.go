package circuit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInvalidCircuitBreaker(t *testing.T) {
	c, err := NewCountCB(0, 1)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "failureThreshold")

	c, err = NewCountCB(1, 0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "halfOpenThreshold")
}

func TestNewvalidCircuitBreakerIsClosed(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)
}

func TestClosedSuccess(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Ok(t))
	assert.Equal(t, result, Succeeded)
	assert.Equal(t, c.State(), Closed)
}

func TestClosedFailureStaysClosed(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)
}

func TestClosedToOpen(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)
}

func TestOpenRejectsCalls(t *testing.T) {
	c, err := NewCountCB(2, 2)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)

	result = c.Call(Ok(t))
	assert.Equal(t, result, Rejected)
	assert.Equal(t, c.State(), Open)
}

func TestOpenToHalfOpen(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)

	result = c.Call(Ok(t))
	assert.Equal(t, result, Rejected)
	assert.Equal(t, c.State(), HalfOpen)
}

func TestHalfOpenSuccessToClosed(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)

	result = c.Call(Ok(t))
	assert.Equal(t, result, Rejected)
	assert.Equal(t, c.State(), HalfOpen)

	result = c.Call(Ok(t))
	assert.Equal(t, result, Succeeded)
	assert.Equal(t, c.State(), Closed)
}

func TestHalfOpenFailureOpen(t *testing.T) {
	c, err := NewCountCB(2, 1)
	assert.NotNil(t, c)
	assert.NoError(t, err)
	assert.Equal(t, c.State(), Closed)

	result := c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Closed)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)

	result = c.Call(Ok(t))
	assert.Equal(t, result, Rejected)
	assert.Equal(t, c.State(), HalfOpen)

	result = c.Call(Error(t))
	assert.Equal(t, result, Failed)
	assert.Equal(t, c.State(), Open)
}
