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

func NewTimeCB(openTimeout time.Duration, halfOpenProbesThreshold, closedFailuresThreshold uint8) (*TimeCB, error) {
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
		clock:                   &RealClock{},
		state:                   Closed,
		openTimeout:             openTimeout,
		openAt:                  nil,
		closedFailures:          0,
		closedFailuresThreshold: closedFailuresThreshold,
		halfOpenProbes:          0,
		halfOpenProbesThreshold: halfOpenProbesThreshold,
	}, nil
}
