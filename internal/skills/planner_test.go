package skills

import (
	"testing"

	"github.com/jonwraymond/toolcompose/skill"
	"github.com/stretchr/testify/require"
)

func TestCompilePlan_EnforcesGuards(t *testing.T) {
	def := skill.Skill{
		Name: "test-skill",
		Steps: []skill.Step{
			{ID: "a", ToolID: "tool:a"},
			{ID: "b", ToolID: "tool:b"},
		},
	}

	_, err := CompilePlan(def, []skill.Guard{skill.MaxStepsGuard(1)})
	require.ErrorIs(t, err, skill.ErrMaxStepsExceeded)

	_, err = CompilePlan(def, []skill.Guard{skill.AllowedToolIDsGuard([]string{"tool:a"})})
	require.ErrorIs(t, err, skill.ErrToolNotAllowed)
}
