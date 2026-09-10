package server

import (
	contextpkg "context"
	"io"

	"github.com/sourcegraph/jsonrpc2"
	"github.com/tliron/glsp"
)

func (s *Server) handle(context contextpkg.Context, connection *jsonrpc2.Conn, request *jsonrpc2.Request) (any, error) {
	if s.Log != nil {
		s.Log.Infof("JSONRPC Handler: Handling request method=%s", request.Method)

		// Log request params for debugging
		if request.Params != nil {
			paramsJSON, _ := request.Params.MarshalJSON()
			s.Log.Infof("JSONRPC Request Params: %s", string(paramsJSON))
		}
	}
	glspContext := glsp.Context{
		Method: request.Method,
		Notify: func(method string, params any) {
			if err := connection.Notify(context, method, params); err != nil {
				s.Log.Errorf("%s", err.Error())
			}
		},
		Call: func(method string, params any, result any) {
			if err := connection.Call(context, method, params, result); err != nil {
				s.Log.Errorf("%s", err.Error())
			}
		},
	}

	if request.Params != nil {
		glspContext.Params = *request.Params
	}

	switch request.Method {
	case "exit":
		// We're giving the attached handler a chance to handle it first, but we'll ignore any result
		_, _, _, _ = s.Handler.Handle(&glspContext)
		err := connection.Close()
		return nil, err

	default:
		// Note: jsonrpc2 will not even call this function if reqest.Params is not valid JSON,
		// so we don't need to handle jsonrpc2.CodeParseError here
		r, validMethod, validParams, err := s.Handler.Handle(&glspContext)
		switch {
		case !validMethod:
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeMethodNotFound,
				Message: "method not supported: " + request.Method,
			}
		case !validParams:
			if err != nil {
				return nil, &jsonrpc2.Error{
					Code:    jsonrpc2.CodeInvalidParams,
					Message: err.Error(),
				}
			} else {
				return nil, &jsonrpc2.Error{
					Code: jsonrpc2.CodeInvalidParams,
				}
			}
		case err != nil:
			return nil, &jsonrpc2.Error{
				Code:    jsonrpc2.CodeInvalidRequest,
				Message: err.Error(),
			}
		default:
			return r, nil
		}
	}
}

func (s *Server) newHandler() jsonrpc2.Handler {
	return jsonrpc2.HandlerWithError(s.handle)
}

func (s *Server) getStreamConn(stream io.ReadWriteCloser) *jsonrpc2.Conn {
	if s.Log != nil {
		s.Log.Info("JSONRPC: Creating new connection with stream")
	}

	// Create a buffered stream with a large buffer size
	bufferedStream := jsonrpc2.NewBufferedStream(stream, jsonrpc2.VSCodeObjectCodec{})

	handler := s.newHandler()
	conn := jsonrpc2.NewConn(s.Context, bufferedStream, handler)

	if s.Log != nil {
		s.Log.Info("JSONRPC: Connection established")
	}

	return conn
}

func (s *Server) ServeStream(stream io.ReadWriteCloser) {
	if s.Log != nil {
		s.Log.Info("🚀 JSONRPC: Starting to serve from stream")
	}
	conn := s.getStreamConn(stream)
	if s.Log != nil {
		s.Log.Info("⏳ JSONRPC: Waiting for disconnect notification")
	}
	<-conn.DisconnectNotify()
	if s.Log != nil {
		s.Log.Info("👋 JSONRPC: Disconnect notification received, stream closed")
	}
}
