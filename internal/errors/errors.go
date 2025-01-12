package errors

import (
	goerrors "errors"
	"fmt"
	wrapper "github.com/go-faster/errors"
)

type DetailedError struct {
	Log  error
	User error
}

func (Det *DetailedError) Error() string {
	return Det.Log.Error()
}

func New(s string) *DetailedError {
	return &DetailedError{
		Log: goerrors.New(s),
	}
}

func Newf(format string, args ...interface{}) *DetailedError {
	return &DetailedError{
		Log: fmt.Errorf(format, args...),
	}
}

func (Det *DetailedError) AddLog(s string) *DetailedError {
	Det.Log = wrapper.Wrap(Det.Log, s)
	return Det
}

func (Det *DetailedError) AddLogf(format string, a ...interface{}) *DetailedError {
	Det.Log = wrapper.Wrapf(Det.Log, format, a)
	return Det
}

func (Det *DetailedError) AddUser(s string) *DetailedError {
	Det.User = wrapper.Wrap(Det.User, s)
	return Det
}

func (Det *DetailedError) AddUserf(format string, a ...interface{}) *DetailedError {
	Det.User = wrapper.Wrapf(Det.User, format, a)
	return Det
}
