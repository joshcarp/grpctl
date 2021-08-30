package grpctl

type Error int

const (
	UnknownError Error = iota
	NotFoundError
)

func (e Error) Error() string {
	switch e {
	case NotFoundError:
		return "not found"
	default:
		return "unknown error"
	}
}