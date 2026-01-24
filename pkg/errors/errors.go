package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeInternalError      Code = "INTERNAL_ERROR"
	CodeInvalidArgument    Code = "INVALID_ARGUMENT"
	CodeNotFound           Code = "NOT_FOUND"
	CodeAlreadyExists      Code = "ALREADY_EXISTS"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeConflict           Code = "CONFLICT"
	CodeUnprocessable      Code = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests    Code = "TOO_MANY_REQUESTS"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeDeadlineExceeded   Code = "DEADLINE_EXCEEDED"
	CodeUnknown            Code = "UNKNOWN"
)

type Error struct {
	Code       Code        `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Internal   error       `json:"-"`
	StackTrace string      `json:"-"`
}

func (e *Error) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Internal
}

func (e *Error) StatusCode() int {
	switch e.Code {
	case CodeInternalError:
		return http.StatusInternalServerError
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Newf(code Code, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func Wrap(err error, code Code, message string) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{
		Code:     code,
		Message:  message,
		Internal: err,
	}
}

func Wrapf(err error, code Code, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:     code,
		Message:  fmt.Sprintf(format, args...),
		Internal: err,
	}
}

func Is(err error, code Code) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

func Equal(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	var e1, e2 *Error
	if errors.As(err1, &e1) && errors.As(err2, &e2) {
		return e1.Code == e2.Code && e1.Message == e2.Message
	}
	return err1.Error() == err2.Error()
}

func NotFound(format string, args ...interface{}) *Error {
	return Newf(CodeNotFound, format, args...)
}

func AlreadyExists(format string, args ...interface{}) *Error {
	return Newf(CodeAlreadyExists, format, args...)
}

func InvalidArgument(format string, args ...interface{}) *Error {
	return Newf(CodeInvalidArgument, format, args...)
}

func Unauthorized(format string, args ...interface{}) *Error {
	return Newf(CodeUnauthorized, format, args...)
}

func Forbidden(format string, args ...interface{}) *Error {
	return Newf(CodeForbidden, format, args...)
}

func Conflict(format string, args ...interface{}) *Error {
	return Newf(CodeConflict, format, args...)
}

func InternalError(format string, args ...interface{}) *Error {
	return Newf(CodeInternalError, format, args...)
}

func TooManyRequests(format string, args ...interface{}) *Error {
	return Newf(CodeTooManyRequests, format, args...)
}

func ServiceUnavailable(format string, args ...interface{}) *Error {
	return Newf(CodeServiceUnavailable, format, args...)
}

type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	return fmt.Sprintf("validation failed with %d errors", len(v))
}

func (v ValidationErrors) Is(target error) bool {
	_, ok := target.(ValidationErrors)
	return ok
}

func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

func NewValidationErrors(errors ...ValidationError) ValidationErrors {
	return errors
}

type Aggregate struct {
	Errors []error
}

func (a *Aggregate) Error() string {
	if len(a.Errors) == 0 {
		return ""
	}
	if len(a.Errors) == 1 {
		return a.Errors[0].Error()
	}
	return fmt.Sprintf("%d errors occurred", len(a.Errors))
}

func (a *Aggregate) Is(target error) bool {
	_, ok := target.(*Aggregate)
	return ok
}

func NewAggregate(errors []error) *Aggregate {
	result := make([]error, 0, len(errors))
	for _, err := range errors {
		if err != nil {
			result = append(result, err)
		}
	}
	return &Aggregate{Errors: result}
}

func (a *Aggregate) Empty() bool {
	return len(a.Errors) == 0
}
