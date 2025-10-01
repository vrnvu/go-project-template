package circuit

import (
	"fmt"
	"time"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c *RealClock) Now() time.Time {
	return time.Now()
}

type TimeCB struct {
	clock                   Clock
	state                   State
	openTimeout             time.Duration
	openAt                  *time.Time
	closedFailures          uint8
	closedFailuresThreshold uint8
	halfOpenProbes          uint8
	halfOpenProbesThreshold uint8
}

func NewTimeCB(clock Clock, openTimeout time.Duration, halfOpenProbesThreshold, closedFailuresThreshold uint8) (*TimeCB, error) {
	if openTimeout > 5*time.Second || openTimeout <= 0*time.Second {
		return nil, fmt.Errorf("openTimeout: 0 < %q < 5", openTimeout)
	}

	if halfOpenProbesThreshold <= 0 {
		return nil, fmt.Errorf("halfOpenProbesThreshold: %q <= 0", halfOpenProbesThreshold)
	}

	if closedFailuresThreshold <= 0 {
		return nil, fmt.Errorf("closedFailuresThreshold: %q <= 0", closedFailuresThreshold)
	}

	return &TimeCB{
		clock:                   clock,
		state:                   Closed,
		openTimeout:             openTimeout,
		openAt:                  nil,
		closedFailures:          0,
		closedFailuresThreshold: closedFailuresThreshold,
		halfOpenProbes:          0,
		halfOpenProbesThreshold: halfOpenProbesThreshold,
	}, nil
}

func (c *TimeCB) Call(f func() error) Result {
	switch c.state {
	case Closed:
		asserts(c.closedFailures < c.closedFailuresThreshold)
		asserts(c.halfOpenProbes == 0)
		asserts(c.openAt == nil)

		if err := f(); err != nil {
			c.closedFailures += 1
			if c.closedFailures == c.closedFailuresThreshold {
				c.state = Open
				now := c.clock.Now()
				c.openAt = &now
			}
			return Failed
		} else {
			c.closedFailures = 0
			return Succeeded
		}
	case Open:
		asserts(c.closedFailures == c.closedFailuresThreshold)
		asserts(c.halfOpenProbes == 0)
		asserts(c.openAt != nil)

		openAtvalue := *c.openAt
		if c.clock.Now().After(openAtvalue.Add(c.openTimeout)) {
			c.state = HalfOpen
			c.halfOpenProbes = 0

			if err := f(); err != nil {
				c.halfOpenProbes += 1
				if c.halfOpenProbes == c.halfOpenProbesThreshold {
					c.state = Open
					c.halfOpenProbes = 0
					now := c.clock.Now()
					c.openAt = &now
				}
				return Failed
			} else {
				c.state = Closed
				c.closedFailures = 0
				c.openAt = nil
				c.halfOpenProbes = 0
				return Succeeded
			}
		} else {
			return Rejected
		}
	case HalfOpen:
		asserts(c.closedFailures == c.closedFailuresThreshold)
		asserts(c.halfOpenProbes < c.halfOpenProbesThreshold)
		asserts(c.openAt != nil)
		openAtValue := *c.openAt
		asserts(c.clock.Now().After(openAtValue.Add(c.openTimeout)))

		if err := f(); err != nil {
			c.halfOpenProbes += 1
			if c.halfOpenProbes == c.halfOpenProbesThreshold {
				c.state = Open
				c.halfOpenProbes = 0
				now := c.clock.Now()
				c.openAt = &now
			}
			return Failed
		} else {
			c.state = Closed
			c.closedFailures = 0
			c.openAt = nil
			c.halfOpenProbes = 0
			return Succeeded
		}
	default:
		panic("unreachable")
	}
}

func (c *TimeCB) State() State {
	return c.state
}
