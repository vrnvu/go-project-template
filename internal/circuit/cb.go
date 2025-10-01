package circuit

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type Result int

const (
	Rejected Result = iota
	Failed
	Succeeded
)

func asserts(condition bool) {
	if !condition {
		panic("assertion failed")
	}
}
