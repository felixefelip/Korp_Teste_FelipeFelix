package msgerr

import "errors"

type poison struct {
	cause error
}

func (p poison) Error() string {
	return "message will never be handled: " + p.cause.Error()
}

func (p poison) Unwrap() error {
	return p.cause
}

func Poison(cause error) error {
	return poison{cause: cause}
}

func IsPoison(err error) bool {
	var target poison

	return errors.As(err, &target)
}
