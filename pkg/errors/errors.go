package errors

import (
	goErrs "errors"
	"fmt"
	"runtime/debug"

	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type Severity string

const (
	SeverityError Severity = "error"
	SeverityWarn  Severity = "warn"
	SeverityHint  Severity = "hint"
)

const (
	ErrGeneration = Error("failed to generate output")
	ErrOpenAPIV2  = Error("openapi v2 is not supported")
)

func IsValidationErr(err error) bool {
	var vErr *ValidationError
	return goErrs.As(err, &vErr)
}

func IsUnsupported(err error) bool {
	var unsupportedErr *UnsupportedError
	return goErrs.As(err, &unsupportedErr)
}

func GetValidationErr(err error) *ValidationError {
	var vErr *ValidationError
	if goErrs.As(err, &vErr) {
		return vErr
	}

	return nil
}

func GetUnsupportedErr(err error) *UnsupportedError {
	var unsupportedErr *UnsupportedError
	if goErrs.As(err, &unsupportedErr) {
		return unsupportedErr
	}

	return nil
}

type ValidationError struct {
	Message  string
	Node     *yaml.Node
	Cause    error
	Severity Severity
	Rule     string
}

func NewValidationError(msg string, node *yaml.Node, err error) *ValidationError {
	return &ValidationError{
		Message:  msg,
		Node:     node,
		Cause:    err,
		Severity: SeverityError,
	}
}

func NewValidationWarning(msg string, node *yaml.Node, err error) *ValidationError {
	return &ValidationError{
		Message:  msg,
		Node:     node,
		Cause:    err,
		Severity: SeverityWarn,
	}
}

func NewValidationHint(msg string, node *yaml.Node, err error) *ValidationError {
	return &ValidationError{
		Message:  msg,
		Node:     node,
		Cause:    err,
		Severity: SeverityHint,
	}
}

// GetLineNumber returns the line number from the Node, or -1 if not available
func (e *ValidationError) GetLineNumber() int {
	if e.Node != nil && e.Node.Line > 0 {
		return e.Node.Line
	}
	return -1
}

// GetColumnNumber returns the column number from the Node, or -1 if not available
func (e *ValidationError) GetColumnNumber() int {
	if e.Node != nil && e.Node.Column > 0 {
		return e.Node.Column
	}
	return -1
}

func NewCircularReferenceError(visited []string) error {
	return goErrs.New(GetCircularReferenceString(visited))
}

func GetCircularReferenceString(visited []string) string {
	loop := ""

	for _, ref := range visited {
		loop = fmt.Sprintf("%s -> %s", loop, ref)
	}

	return loop
}

func NewResolveError(err error, node *yaml.Node) *ValidationError {
	return &ValidationError{
		Message:  "error resolving schema",
		Node:     node,
		Cause:    err,
		Severity: SeverityError,
	}
}

func (e ValidationError) Error() string {
	lineStr := ""
	lineNum := e.GetLineNumber()
	colNum := e.GetColumnNumber()
	if lineNum > 0 {
		if colNum > 0 {
			lineStr = fmt.Sprintf("[line %d:%d] ", lineNum, colNum)
		} else {
			lineStr = fmt.Sprintf("[line %d] ", lineNum)
		}
	}
	errorSuffix := ""
	if e.Cause != nil {
		errorSuffix = ": " + e.Cause.Error()
	}
	pathSuffix := ""
	rulePrefix := ""
	if e.Rule != "" {
		rulePrefix = e.Rule + " - "
	}

	message := fmt.Sprintf("%s%s%s", rulePrefix, e.Message, pathSuffix)

	return fmt.Sprintf("validation %s: %s%s%s", e.Severity, lineStr, message, errorSuffix)
}

func (e ValidationError) Is(target error) bool {
	validationErr, ok := target.(*ValidationError)
	if !ok {
		return goErrs.Is(e.Cause, target)
	}

	return goErrs.Is(e.Cause, validationErr.Cause)
}

func (e ValidationError) Unwrap() error {
	return e.Cause
}

func (e ValidationError) GetResult() *validation.Error {
	errorSuffix := ""
	if e.Cause != nil {
		errorSuffix = ": " + e.Cause.Error()
	}

	message := fmt.Sprintf("%s%s", e.Message, errorSuffix)

	// Map severity to validation.Severity
	var sev validation.Severity
	switch e.Severity {
	case SeverityError:
		sev = validation.SeverityError
	case SeverityWarn:
		sev = validation.SeverityWarning
	case SeverityHint:
		sev = validation.SeverityHint
	default:
		sev = validation.SeverityError
	}

	// Use rule ID if set, otherwise use "general"
	ruleID := e.Rule
	if ruleID == "" {
		ruleID = "general"
	}

	// Get the node or create a synthetic one with line number
	node := e.Node
	if node == nil {
		lineNum := e.GetLineNumber()
		if lineNum < 0 {
			lineNum = 0
		}
		node = &yaml.Node{
			Line: lineNum,
		}
	}

	return &validation.Error{
		Rule:            ruleID,
		Severity:        sev,
		Node:            node,
		UnderlyingError: goErrs.New(message),
	}
}

type UnsupportedError struct {
	Message string
	Node    *yaml.Node
}

func NewUnsupportedError(msg string, node *yaml.Node) *UnsupportedError {
	return &UnsupportedError{
		Message: msg,
		Node:    node,
	}
}

func (e UnsupportedError) GetLineNumber() int {
	if e.Node != nil && e.Node.Line > 0 {
		return e.Node.Line
	}
	return -1
}

func (e UnsupportedError) GetColumnNumber() int {
	if e.Node != nil && e.Node.Column > 0 {
		return e.Node.Column
	}
	return -1
}

func (e UnsupportedError) Error() string {
	lineStr := ""
	if e.Node != nil && e.Node.Line > 0 {
		lineStr = fmt.Sprintf(" (line %d)", e.Node.Line)
	}

	return fmt.Sprintf("unsupported: %s%s", e.Message, lineStr)
}

func (e UnsupportedError) Is(target error) bool {
	err, ok := target.(*UnsupportedError)
	if !ok {
		return false
	}

	return e.Message == err.Message && e.Node != nil && err.Node != nil && e.Node.Line == err.Node.Line
}

type SkippedEntity struct {
	Name string
	Type string
}
type SkippedError struct {
	SkippedEntity SkippedEntity
	*UnsupportedError
}

func NewSkippedError(skippedEntity SkippedEntity, uErr *UnsupportedError) *SkippedError {
	return &SkippedError{
		SkippedEntity:    skippedEntity,
		UnsupportedError: uErr,
	}
}

func (e SkippedError) Error() string {
	// SkippedError is solely used for encapsulating UnsupportedError with the skipped entity,
	// we don't want to print the skipped entity in the error message.
	return e.UnsupportedError.Error()
}

type PanicError struct {
	StackTrace string
	Cause      error
}

func NewPanicError(cause error) *PanicError {
	return &PanicError{
		StackTrace: string(debug.Stack()),
		Cause:      cause,
	}
}

func (e PanicError) Error() string {
	return fmt.Sprintf("panic: %s\n%s", e.Cause.Error(), e.StackTrace)
}

func (e PanicError) Is(target error) bool {
	validationErr, ok := target.(*PanicError)
	if !ok {
		return goErrs.Is(e.Cause, target)
	}

	return goErrs.Is(e.Cause, validationErr.Cause) && e.StackTrace == validationErr.StackTrace
}

func (e PanicError) Unwrap() error {
	return e.Cause
}

// New creates a new error with the given message
func New(message string) error {
	return goErrs.New(message)
}

// As will set target errors value to equal Error if they are equivalent
func As(err error, target any) bool {
	return goErrs.As(err, target)
}

// Is checks if err is equivalent to target
func Is(err error, target error) bool {
	return goErrs.Is(err, target)
}

// Join joins the errors together into a single error
func Join(errs ...error) error {
	return goErrs.Join(errs...)
}
