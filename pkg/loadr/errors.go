package loadr

func (e *Error) Error() string {
	return e.Message
}
