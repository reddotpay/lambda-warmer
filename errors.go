package warmer

// Error defines a Lambda warmer error
type Error struct {
	Code string
}

const (
	// ErrCodeNotWarmerEvent defines error code for not warmer event
	ErrCodeNotWarmerEvent = "NotWarmerEvent"
)

var errMessage = map[string]string{
	ErrCodeNotWarmerEvent: "not a lambda warmer event",
}

// New initialises a new Error
func New(code string) *Error {
	return &Error{Code: code}
}

func (err Error) Error() string {
	return errMessage[err.Code]
}
