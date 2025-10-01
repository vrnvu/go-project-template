package circuit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestClock struct {
	now        time.Time
	tickAmount time.Duration
}

func NewTestClock(now time.Time, tickAmount time.Duration) *TestClock {
	return &TestClock{now: now, tickAmount: tickAmount}
}

func (c *TestClock) Now() time.Time {
	return c.now
}

func (c *TestClock) Tick() {
	c.now = c.now.Add(c.tickAmount)
}

func TestNewTimeCBInvalid(t *testing.T) {
	t.Parallel()

	c, err := NewTimeCB(NewTestClock(time.Now(), time.Millisecond), 0*time.Second, 1, 1)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "openTimeout")

	c, err = NewTimeCB(NewTestClock(time.Now(), time.Millisecond), 1*time.Second, 0, 1)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "halfOpenProbesThreshold")

	c, err = NewTimeCB(NewTestClock(time.Now(), time.Millisecond), 1*time.Second, 1, 0)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "closedFailuresThreshold")

	c, err = NewTimeCB(NewTestClock(time.Now(), time.Millisecond), 6*time.Second, 1, 1)
	assert.Nil(t, c)
	assert.ErrorContains(t, err, "openTimeout")
}

func TestTimeClosedSuccess(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Ok(t))
	assert.Equal(t, Succeeded, result)
	assert.Equal(t, Closed, cb.State())
}

func TestTimeClosedFailureStaysClosed(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())
}

func TestTimeClosedToOpen(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())
}

func TestTimeOpenRejectsCallsImmediately(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	result = cb.Call(Ok(t))
	assert.Equal(t, Rejected, result)
	assert.Equal(t, Open, cb.State())
}

func TestTimeOpenRejectsUntilTimeoutThenAllowsHalfOpenCall(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	result = cb.Call(Ok(t))
	assert.Equal(t, Rejected, result)
	assert.Equal(t, Open, cb.State())

	clock.Tick()

	result = cb.Call(Ok(t))
	assert.Equal(t, Succeeded, result)
	assert.Equal(t, Closed, cb.State())
}

func TestTimeHalfOpenSuccessClosesBreaker(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	result = cb.Call(Ok(t))
	assert.Equal(t, Rejected, result)
	assert.Equal(t, Open, cb.State())

	clock.Tick()

	result = cb.Call(Ok(t))
	assert.Equal(t, Succeeded, result)
	assert.Equal(t, Closed, cb.State())
}

func TestTimeHalfOpenFailureOpensBreaker(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	openTimeout := time.Millisecond
	halfOpenProbesThreshold := uint8(1)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, openTimeout, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())
}

func TestTimeHalfOpenRespectsProbesThreshold(t *testing.T) {
	t.Parallel()

	start := time.Now()
	clock := NewTestClock(start, 2*time.Millisecond)
	halfOpenProbesThreshold := uint8(2)
	closedFailuresThreshold := uint8(2)

	cb, err := NewTimeCB(clock, time.Millisecond, halfOpenProbesThreshold, closedFailuresThreshold)
	assert.NoError(t, err)
	assert.Equal(t, Closed, cb.State())

	result := cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Closed, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	clock.Tick()

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, HalfOpen, cb.State())

	result = cb.Call(Error(t))
	assert.Equal(t, Failed, result)
	assert.Equal(t, Open, cb.State())

	result = cb.Call(Ok(t))
	assert.Equal(t, Rejected, result)
	assert.Equal(t, Open, cb.State())
}
