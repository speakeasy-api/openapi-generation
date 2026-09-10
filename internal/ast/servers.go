package ast

import (
	"fmt"
	"slices"
	"strings"
)

type ServerVariable struct {
	Name        string   `yaml:",omitempty"`
	Type        *TypeDef `yaml:",omitempty"`
	Default     string   `yaml:",omitempty"`
	ServerIndex int      `yaml:",omitempty"` // Index of the server this variable belongs to
	Server      *Server  `yaml:",omitempty"` // Reference to the server this variable belongs to
}

func (v *ServerVariable) Match(matchers Matchers) error {
	if matchers.ServerVariable != nil {
		return matchers.ServerVariable(v)
	}

	return nil
}

// Server represents a single server that is available to an SDK or operation
type Server struct {
	ID         string            `yaml:",omitempty"` // The ID of the server (if any)
	URL        string            `yaml:",omitempty"` // Server URLs can be templated strings containing braces, e.g. "https://{env}.example.com"
	IsRelative bool              `yaml:",omitempty"` // Whether or not the url is relative
	Comments   *Comment          `yaml:",omitempty"` // The comments associated with the server
	Variables  []*ServerVariable `yaml:",omitempty"` // The variables associated with the url if it is templated
}

func (s *Server) Match(matchers Matchers) error {
	if matchers.Server != nil {
		return matchers.Server(s)
	}

	return nil
}

// Servers represents the list of servers that are available to and SDK or operation
type Servers struct {
	Servers   []*Server `yaml:",omitempty"` // The list of servers
	Default   string    `yaml:",omitempty"` // The ID of the default server (if any)
	ServerMap bool      `yaml:",omitempty"` // Whether or not the list of servers should be represented as a map
}

func (s *Servers) Match(matchers Matchers) error {
	if matchers.Servers != nil {
		return matchers.Servers(s)
	}

	return nil
}

// GetDefaultURL returns the URL of the default server
func (s *Servers) GetDefaultURL(useDefaultVariables bool) string {
	url := ""

	for _, server := range s.Servers {
		if server.ID == s.Default && !server.IsRelative {
			url = server.URL
			if useDefaultVariables {
				variables := s.GetVariables()
				for _, variable := range variables {
					varStr := fmt.Sprintf("{%s}", variable.Name)
					url = strings.ReplaceAll(url, varStr, variable.Default)
				}
			}
		}
	}
	return url
}

// GetVariables returns the variables associated with the servers
func (s *Servers) GetVariables() []*ServerVariable {
	variables := []*ServerVariable{}

	for serverIndex, server := range s.Servers {
		for _, variable := range server.Variables {
			// Create a copy with server correlation info
			varWithServerInfo := &ServerVariable{
				Name:        variable.Name,
				Type:        variable.Type,
				Default:     variable.Default,
				ServerIndex: serverIndex,
				Server:      server,
			}

			idx := slices.IndexFunc(variables, func(v *ServerVariable) bool {
				return v.Name == variable.Name
			})
			if idx == -1 {
				variables = append(variables, varWithServerInfo)
			} else {
				// If variable already exists, keep the first occurrence but update server info
				// This handles cases where the same variable appears in multiple servers
				variables[idx] = varWithServerInfo
			}
		}
	}

	return variables
}

// HasAbsoluteURL returns true if the servers list contains at least one server with an absolute URL
func (s *Servers) HasAbsoluteURL() bool {
	for _, server := range s.Servers {
		if !server.IsRelative {
			return true
		}
	}

	return false
}
