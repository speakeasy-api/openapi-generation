package models

type Error struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Type    string      `json:"type"`
	Cause   *ErrorCause `json:"cause,omitempty"`
}

type ErrorCause struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}

type HeaderAuth struct {
	HeaderName    string `json:"headerName"`
	ExpectedValue string `json:"expectedValue"`
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthRequest struct {
	HeaderAuth []HeaderAuth `json:"headerAuth,omitempty"`
	BasicAuth  *BasicAuth   `json:"basicAuth,omitempty"`
}

type CustomHttpAuthResponse struct {
	Grant  string   `json:"grant"`
	Scopes []string `json:"scopes"`
}
