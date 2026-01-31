package auth

import (
	"context"
	"slices"
	"strings"
)

// RBACConfig configures the simple RBAC authorizer.
type RBACConfig struct {
	// Roles maps role names to their configuration.
	Roles map[string]RoleConfig

	// DefaultRole is used when the subject has no roles.
	DefaultRole string
}

// RoleConfig defines permissions for a role.
type RoleConfig struct {
	// Permissions are explicit permission strings.
	// Use "*" for wildcard (all permissions).
	Permissions []string

	// Inherits lists roles whose permissions are inherited.
	Inherits []string

	// AllowedTools are tool name patterns this role can access.
	// Supports "*" (all), "prefix*", and "*suffix" patterns.
	AllowedTools []string

	// DeniedTools are tool patterns explicitly denied.
	// Takes precedence over AllowedTools.
	DeniedTools []string

	// AllowedActions are actions this role can perform.
	// If empty, all actions are allowed (when tool is allowed).
	AllowedActions []string
}

// SimpleRBACAuthorizer implements role-based access control.
type SimpleRBACAuthorizer struct {
	config RBACConfig
}

// NewSimpleRBACAuthorizer creates a new RBAC authorizer.
func NewSimpleRBACAuthorizer(config RBACConfig) *SimpleRBACAuthorizer {
	return &SimpleRBACAuthorizer{config: config}
}

// Name returns "simple_rbac".
func (a *SimpleRBACAuthorizer) Name() string {
	return "simple_rbac"
}

// Authorize checks if the subject is allowed to perform the action.
func (a *SimpleRBACAuthorizer) Authorize(_ context.Context, req *AuthzRequest) error {
	roles := a.getEffectiveRoles(req.Subject)

	if len(roles) == 0 && a.config.DefaultRole != "" {
		roles = []string{a.config.DefaultRole}
	}

	toolName := req.ToolName()

	// Check each role for permission
	for _, roleName := range roles {
		if a.roleAllows(roleName, toolName, req.Action, make(map[string]bool)) {
			return nil
		}
	}

	// Build denial error
	subject := ""
	if req.Subject != nil {
		subject = req.Subject.Principal
	}

	return &AuthzError{
		Subject:  subject,
		Resource: req.Resource,
		Action:   req.Action,
		Reason:   "no role grants permission",
	}
}

func (a *SimpleRBACAuthorizer) getEffectiveRoles(subject *Identity) []string {
	if subject == nil {
		return nil
	}
	return subject.Roles
}

func (a *SimpleRBACAuthorizer) roleAllows(roleName, tool, action string, visited map[string]bool) bool {
	// Prevent infinite recursion from circular inheritance
	if visited[roleName] {
		return false
	}
	visited[roleName] = true

	role, ok := a.config.Roles[roleName]
	if !ok {
		return false
	}

	// Check wildcard permission
	if slices.Contains(role.Permissions, "*") {
		return true
	}

	// Check denied tools first (deny takes precedence)
	for _, pattern := range role.DeniedTools {
		if matchPattern(tool, pattern) {
			return false
		}
	}

	// Check allowed tools
	toolAllowed := len(role.AllowedTools) == 0 // If no tools specified, check inherited
	for _, pattern := range role.AllowedTools {
		if matchPattern(tool, pattern) {
			toolAllowed = true
			break
		}
	}

	// Check allowed actions
	actionAllowed := len(role.AllowedActions) == 0 // If no actions specified, all are allowed
	for _, allowedAction := range role.AllowedActions {
		if allowedAction == action || allowedAction == "*" {
			actionAllowed = true
			break
		}
	}

	// If this role grants permission, return true
	if toolAllowed && actionAllowed {
		// Only if at least one of tools or actions was explicitly specified
		if len(role.AllowedTools) > 0 || len(role.AllowedActions) > 0 || len(role.Permissions) > 0 {
			return true
		}
	}

	// Check inherited roles
	for _, inherited := range role.Inherits {
		if a.roleAllows(inherited, tool, action, visited) {
			return true
		}
	}

	return false
}

// matchPattern matches a value against a pattern.
// Supports: "*" (match all), "prefix*", "*suffix", exact match.
func matchPattern(value, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Check for *contains* pattern (e.g., "*foo*")
	if len(pattern) > 2 && pattern[0] == '*' && pattern[len(pattern)-1] == '*' {
		// Contains match - not fully supported, treat as substring search
		middle := pattern[1 : len(pattern)-1]
		return strings.Contains(value, middle)
	}

	if prefix, found := strings.CutSuffix(pattern, "*"); found {
		return strings.HasPrefix(value, prefix)
	}

	if suffix, found := strings.CutPrefix(pattern, "*"); found {
		return strings.HasSuffix(value, suffix)
	}

	return value == pattern
}
