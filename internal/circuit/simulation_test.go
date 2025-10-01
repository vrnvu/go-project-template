package circuit

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StepCount int

const (
	CountSuccess StepCount = iota
	CountFailure
)

type StepTime int

const (
	TimeSuccess StepTime = iota
	TimeFailure
	TimeTick
)

func generateRandomStepsCount(t *testing.T, seed int64, count int) []StepCount {
	t.Helper()

	steps := make([]StepCount, 0, count)
	r := rand.New(rand.NewSource(seed)) //nolint:gosec
	for range count {
		if r.Int()%2 == 0 {
			steps = append(steps, CountSuccess)
		} else {
			steps = append(steps, CountFailure)
		}
	}

	return steps
}

func TestCountCBRandomeSequence(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("slow/integration: count sim")
	}

	failureThreshold := 10
	halfOpenThreshold := 4
	seed := int64(42)
	count := 100_000
	steps := generateRandomStepsCount(t, seed, count)

	cb, err := NewCountCB(uint8(failureThreshold), uint8(halfOpenThreshold))
	assert.NotNil(t, cb)
	assert.NoError(t, err)

	for _, step := range steps {
		switch step {
		case CountSuccess:
			assert.NotPanics(t, func() {
				_ = cb.Call(Ok(t))
			})
		case CountFailure:
			assert.NotPanics(t, func() {
				_ = cb.Call(Error(t))
			})
		default:
			panic("unreachable")
		}
	}
}
