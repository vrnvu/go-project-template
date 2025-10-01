package circuit

import (
	"fmt"
)

type CountCB struct {
	state                   State
	closedFailures          uint8
	closedFailuresThreshold uint8
	halfOpenAttempts        uint8
	halfOpenThreshold       uint8
}

func NewCountCB(failureTreshold, halfOpenThreshold uint8) (*CountCB, error) {
	if failureTreshold <= 0 {
		return nil, fmt.Errorf("failureThreshold: %q <= 0", failureTreshold)
	}

	if halfOpenThreshold <= 0 {
		return nil, fmt.Errorf("halfOpenThreshold: %q <= 0", failureTreshold)
	}

	return &CountCB{
		state:                   Closed,
		closedFailures:          0,
		closedFailuresThreshold: failureTreshold,
		halfOpenAttempts:        0,
		halfOpenThreshold:       halfOpenThreshold,
	}, nil
}

func (c *CountCB) Call(f func() error) Result {
	switch c.State() {
	case Closed:
		asserts(c.closedFailures < c.closedFailuresThreshold)
		asserts(c.halfOpenAttempts == 0)

		if err := f(); err != nil {
			c.closedFailures += 1
			if c.closedFailures == c.closedFailuresThreshold {
				c.state = Open
			}
			return Failed
		} else {
			c.closedFailures = 0
			return Succeeded
		}
	case Open:
		asserts(c.closedFailures == c.closedFailuresThreshold)
		asserts(c.halfOpenAttempts < c.halfOpenThreshold)

		c.halfOpenAttempts += 1
		if c.halfOpenAttempts == c.halfOpenThreshold {
			c.state = HalfOpen
			c.halfOpenAttempts = 0
		}
		return Rejected
	case HalfOpen:
		asserts(c.closedFailures == c.closedFailuresThreshold)
		asserts(c.halfOpenAttempts < c.halfOpenThreshold)

		if err := f(); err != nil {
			c.state = Open
			c.halfOpenAttempts = 0
			return Failed
		} else {
			c.state = Closed
			c.closedFailures = 0
			return Succeeded
		}
	default:
		panic("unreachable")
	}
}

func (c *CountCB) State() State {
	return c.state
}
