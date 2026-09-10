package ast

// Describes an assertion on a response body.
type ResponseBodyAssertion struct {
	Path    string
	Value   *Example
	Content *ResponseBodyContent
}
