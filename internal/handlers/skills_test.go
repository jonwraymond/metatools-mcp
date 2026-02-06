package handlers

import (
	"context"
	"errors"
	"testing"

	internalskills "github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/internal/toolset"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolcompose/skill"
	"github.com/stretchr/testify/require"
)

type stubSkillRunner struct {
	errTool string
}

func (s stubSkillRunner) Run(_ context.Context, toolID string, _ map[string]any) (RunResult, error) {
	if toolID == s.errTool {
		return RunResult{}, errors.New("boom")
	}
	return RunResult{Structured: toolID}, nil
}

func (s stubSkillRunner) RunChain(_ context.Context, _ []ChainStep) (RunResult, []StepResult, error) {
	return RunResult{}, nil, nil
}

func TestSkillsHandler_PlanAndRun(t *testing.T) {
	reg := internalskills.NewRegistry([]*internalskills.Skill{{
		ID:   "skill:demo",
		Name: "demo",
		Steps: []skill.Step{
			{ID: "b", ToolID: "tool:b"},
			{ID: "a", ToolID: "tool:a"},
		},
	}})

	handler := NewSkillsHandler(reg, toolset.NewRegistry(nil), stubSkillRunner{errTool: "tool:b"}, SkillDefaults{
		MaxSteps:     0,
		MaxToolCalls: 0,
	})

	plan, err := handler.Plan(context.Background(), metatools.PlanSkillInput{SkillID: "skill:demo"})
	require.NoError(t, err)
	require.Len(t, plan.Plan.Steps, 2)
	require.Equal(t, "a", plan.Plan.Steps[0].ID)

	out, isError, err := handler.Run(context.Background(), metatools.RunSkillInput{SkillID: "skill:demo"})
	require.NoError(t, err)
	require.True(t, isError)
	require.Len(t, out.Results, 2)
	require.NotNil(t, out.Error)
	require.Equal(t, "a", out.Results[0].StepID)
	require.Equal(t, "tool:a", out.Results[0].Value)
	require.Equal(t, "b", out.Results[1].StepID)
	require.NotNil(t, out.Results[1].Error)
}
