package expl

import (
	"fmt"
	"strings"

	"github.com/srvc/fail"
)

type ErrorCode int

const (
	SQLParseError ErrorCode = iota + 1
	ExeExplainError
	ShowTablesError
	UserInputError
	OtherError
)

var errorMessage = map[ErrorCode]string{
	SQLParseError:   "sql parse error",
	ExeExplainError: "explain error",
	ShowTablesError: "execute show tables sql error",
	UserInputError:  "invalid parameter you inputed",
	OtherError:      "error",
}

func ErrWrap(err error, code ErrorCode) error {
	return fail.Wrap(
		err,
		fail.WithCode(int(code)),
		fail.WithIgnorable(),
	)
}

func ErrWrapWithMessage(err error, code ErrorCode, msg string) error {
	return fail.Wrap(
		err,
		fail.WithCode(int(code)),
		fail.WithIgnorable(),
		fail.WithMessage(msg),
	)
}

func Message(err error) string {
	msg := strings.Join(fail.Unwrap(err).Messages, "\n")

	return fmt.Sprintf("%s %s\nStackTrace:%+v\n",
		errorMessage[ErrorCode(ErrCode(err))],
		msg,
		fail.Unwrap(err).StackTrace,
	)
}

func LogMessage(err error) string {
	return fmt.Sprintf("%T\nCode:%d\nStackTrace:%+v\n",
		err,
		fail.Unwrap(err).Code,
		fail.Unwrap(err).StackTrace,
	)
}

func ErrCode(err error) int {
	return fail.Unwrap(err).Code.(int)
}
