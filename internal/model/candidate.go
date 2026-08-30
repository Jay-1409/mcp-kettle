package model

import (
	"fmt"
	"strings"
)

type Parameter struct {
	Name     string
	Location string
	Type     string
	Required bool
}

type Candidate struct {
	ID          string
	ToolName    string
	Description string
	Method      string
	Route       string
	SourceFile  string
	SourceLine  int
	Parameters  []Parameter
	Issue       string
}

func (c Candidate) Ready() bool { return c.Issue == "" }

func (c Candidate) Label() string {
	return fmt.Sprintf("%-6s %-28s %s:%d", c.Method, c.Route, c.SourceFile, c.SourceLine)
}

func ToolName(method, route string) string {
	name := strings.ToLower(method) + "_" + strings.Trim(route, "/")
	var b strings.Builder
	lastUnderscore := false
	for _, r := range name {
		valid := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if valid {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
