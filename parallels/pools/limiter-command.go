package pools

type Command int

const (
	StopAndJoin Command = iota
	StopAndDetach
	ResetQueue
)
