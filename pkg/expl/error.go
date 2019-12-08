package expl

import (
	"fmt"
	"github.com/srvc/fail"
	"strconv"

	"github.com/morikuni/failure"
)

// ErrorCode error code
type ErrorCode int

const (
	// SQLParseError error of sql parse
	SQLParseError ErrorCode = iota + 1

	// ExeExplainError execute explain error
	ExeExplainError

	// ShowTablesError error in showing tables
	ShowTablesError

	// UserInputError error of input parameter
	UserInputError

	// OtherError error of other reason
	OtherError
)

var errorMessage = map[ErrorCode]string{
	SQLParseError:   "sql parse error",
	ExeExplainError: "explain error",
	ShowTablesError: "execute show tables sql error",
	UserInputError:  "invalid parameter you inputed",
	OtherError:      "error",
}

// ErrWrap wrapping error
func ErrWrap(err error, code ErrorCode) error {
	cd := failure.StringCode(fmt.Sprintf("%d", code))
	err = failure.Wrap(err, failure.WithCode(cd))

	return err
}

// ErrWrapWithMessage wrapping error with message
func ErrWrapWithMessage(err error, code ErrorCode, msg string) error {
	cd := failure.StringCode(fmt.Sprintf("%d", code))
	err = failure.Wrap(err, failure.WithCode(cd), failure.Messagef(msg))

	return err
}

// Message get error message
func Message(err error) string {
	return fmt.Sprintf("%+v", err)
}

// LogMessage get error message for logging
func LogMessage(err error) string {
	code := ErrCode(err)

	return fmt.Sprintf("Code:%d, %s\nStackTrace:%+v\n",
		code,
		errorMessage[ErrorCode(code)],
		fail.Unwrap(err).StackTrace,
	)
}

// ErrCode get error code
func ErrCode(err error) int {
	var code int
	codeVal, ok := failure.CodeOf(err)
	if ok {
		code, _ = strconv.Atoi(codeVal.ErrorCode())
	}
	return code
}
