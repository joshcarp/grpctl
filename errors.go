package grpctl

type Error int

const (
	NotFoundError Error = iota
	AlreadyExists
	InvalidArg
	ContextError
)

func (e Error) Error() string {
	switch e {
	case NotFoundError:
		return "not found"
	case AlreadyExists:
		return "item already exists"
	case InvalidArg:
		return "invalid argument"
	case ContextError:
		return "context not found, must run grpctl through grpctl.RunCommand"
	default:
		return "unknown error"
	}
}
