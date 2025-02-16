package status

type Status int

const (
	Unknown Status = iota - 1
	Success
	Failure
	Warning
)

func (r Status) String() string {
	switch r {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	case Warning:
		return "Warning"
	default:
		return "UNKNOWN"
	}
}
