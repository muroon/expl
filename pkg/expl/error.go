package expl

import (
	"fmt"
	"github.com/srvc/fail"
	"strconv"

	"github.com/morikuni/failure"
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
	cd := failure.StringCode(fmt.Sprintf("%d", code))
	err = failure.Wrap(err, failure.WithCode(cd))

	return err
}

func ErrWrapWithMessage(err error, code ErrorCode, msg string) error {
	cd := failure.StringCode(fmt.Sprintf("%d", code))
	err = failure.Wrap(err, failure.WithCode(cd), failure.Messagef(msg))

	return err
}

func Message(err error) string {
	return fmt.Sprintf("%+v", err)
}

func LogMessage(err error) string {
	code := ErrCode(err)

	return fmt.Sprintf("Code:%d, %s\nStackTrace:%+v\n",
		code,
		errorMessage[ErrorCode(code)],
		fail.Unwrap(err).StackTrace,
	)
}

func ErrCode(err error) int {
	var code int
	codeVal, ok := failure.CodeOf(err)
	if ok {
		code, _ = strconv.Atoi(codeVal.ErrorCode())
	}
	return code
}
