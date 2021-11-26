package grpctl

type Error int

const (
	NotFoundError  Error = iota
	AlreadyExists
	InvalidArg
)

func (e Error) Error() string {
	switch e {
	case NotFoundError:
		return "not found"
	case AlreadyExists:
		return "item already exists"
	case InvalidArg:
		return "invalid argument"
	default:
		return "unknown error"
	}
}
