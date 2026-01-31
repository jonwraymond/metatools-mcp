package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleRBACAuthorizer_AdminRole(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"admin": {
				Permissions: []string{"*"},
			},
		},
	})

	req := &AuthzRequest{
		Subject:  &Identity{Principal: "alice", Roles: []string{"admin"}},
		Resource: "tool:execute_code",
		Action:   "call",
	}

	err := authz.Authorize(context.Background(), req)
	assert.NoError(t, err)
}

func TestSimpleRBACAuthorizer_AllowedTools(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"user": {
				AllowedTools: []string{"search_*", "describe_*"},
			},
		},
	})

	tests := []struct {
		name     string
		tool     string
		expected bool
	}{
		{"allowed search", "search_tools", true},
		{"allowed describe", "describe_tool", true},
		{"denied execute", "execute_code", false},
		{"denied run", "run_tool", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AuthzRequest{
				Subject:  &Identity{Roles: []string{"user"}},
				Resource: "tool:" + tt.tool,
				Action:   "call",
			}
			err := authz.Authorize(context.Background(), req)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSimpleRBACAuthorizer_DeniedAction(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"reader": {
				AllowedActions: []string{"list", "describe"},
			},
		},
	})

	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{"allowed list", "list", true},
		{"allowed describe", "describe", true},
		{"denied call", "call", false},
		{"denied execute", "execute", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AuthzRequest{
				Subject:  &Identity{Roles: []string{"reader"}},
				Resource: "tool:some_tool",
				Action:   tt.action,
			}
			err := authz.Authorize(context.Background(), req)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSimpleRBACAuthorizer_DeniedToolsPrecedence(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"user": {
				AllowedTools: []string{"*"},          // Allow all
				DeniedTools:  []string{"execute_*"}, // But deny execute_*
			},
		},
	})

	tests := []struct {
		name     string
		tool     string
		expected bool
	}{
		{"allowed search", "search_tools", true},
		{"denied execute_code", "execute_code", false},
		{"denied execute_script", "execute_script", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AuthzRequest{
				Subject:  &Identity{Roles: []string{"user"}},
				Resource: "tool:" + tt.tool,
				Action:   "call",
			}
			err := authz.Authorize(context.Background(), req)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSimpleRBACAuthorizer_RoleInheritance(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"base": {
				AllowedTools: []string{"search_*"},
			},
			"advanced": {
				Inherits:     []string{"base"},
				AllowedTools: []string{"run_*"},
			},
		},
	})

	req := &AuthzRequest{
		Subject:  &Identity{Roles: []string{"advanced"}},
		Resource: "tool:search_tools",
		Action:   "call",
	}

	// Should inherit base permissions
	err := authz.Authorize(context.Background(), req)
	assert.NoError(t, err)

	// Should also have own permissions
	req.Resource = "tool:run_tool"
	err = authz.Authorize(context.Background(), req)
	assert.NoError(t, err)
}

func TestSimpleRBACAuthorizer_DefaultRole(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		DefaultRole: "anonymous",
		Roles: map[string]RoleConfig{
			"anonymous": {
				AllowedTools:   []string{"search_*"},
				AllowedActions: []string{"list"},
			},
		},
	})

	// Identity with no roles falls back to default
	req := &AuthzRequest{
		Subject:  &Identity{Roles: nil},
		Resource: "tool:search_tools",
		Action:   "list",
	}

	err := authz.Authorize(context.Background(), req)
	assert.NoError(t, err)
}

func TestSimpleRBACAuthorizer_NoMatchingRole(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"admin": {
				Permissions: []string{"*"},
			},
		},
	})

	req := &AuthzRequest{
		Subject:  &Identity{Roles: []string{"unknown_role"}},
		Resource: "tool:anything",
		Action:   "call",
	}

	err := authz.Authorize(context.Background(), req)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		value    string
		expected bool
	}{
		{"*", "anything", true},
		{"search_*", "search_tools", true},
		{"search_*", "run_tool", false},
		{"*_tool", "run_tool", true},
		{"*_tool", "search_tools", false},
		{"exact", "exact", true},
		{"exact", "different", false},
		{"", "", true},
		{"", "something", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.value, func(t *testing.T) {
			assert.Equal(t, tt.expected, matchPattern(tt.value, tt.pattern))
		})
	}
}

func TestSimpleRBACAuthorizer_Name(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{})
	assert.Equal(t, "simple_rbac", authz.Name())
}

func TestSimpleRBACAuthorizer_NilSubject(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		DefaultRole: "anonymous",
		Roles: map[string]RoleConfig{
			"anonymous": {
				AllowedTools: []string{"*"},
			},
		},
	})

	req := &AuthzRequest{
		Subject:  nil,
		Resource: "tool:anything",
		Action:   "call",
	}

	// Should use default role when subject is nil
	err := authz.Authorize(context.Background(), req)
	require.NoError(t, err)
}

func TestSimpleRBACAuthorizer_MultipleRoles(t *testing.T) {
	authz := NewSimpleRBACAuthorizer(RBACConfig{
		Roles: map[string]RoleConfig{
			"viewer": {
				AllowedActions: []string{"list"},
			},
			"executor": {
				AllowedActions: []string{"call"},
			},
		},
	})

	// User with both roles
	req := &AuthzRequest{
		Subject:  &Identity{Roles: []string{"viewer", "executor"}},
		Resource: "tool:anything",
		Action:   "call",
	}

	err := authz.Authorize(context.Background(), req)
	assert.NoError(t, err)

	req.Action = "list"
	err = authz.Authorize(context.Background(), req)
	assert.NoError(t, err)
}
