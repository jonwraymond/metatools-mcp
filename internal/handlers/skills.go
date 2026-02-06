package handlers

import (
	"context"
	"errors"
	"time"

	merrors "github.com/jonwraymond/metatools-mcp/internal/errors"
	internalskills "github.com/jonwraymond/metatools-mcp/internal/skills"
	"github.com/jonwraymond/metatools-mcp/pkg/metatools"
	"github.com/jonwraymond/toolcompose/skill"
)

// SkillsHandler handles skill metatools.
type SkillsHandler struct {
	registry SkillRegistry
	toolsets ToolsetRegistry
	runner   Runner
	defaults SkillDefaults
}

// NewSkillsHandler creates a new skills handler.
func NewSkillsHandler(registry SkillRegistry, toolsets ToolsetRegistry, runner Runner, defaults SkillDefaults) *SkillsHandler {
	return &SkillsHandler{
		registry: registry,
		toolsets: toolsets,
		runner:   runner,
		defaults: defaults,
	}
}

// List handles list_skills.
func (h *SkillsHandler) List(ctx context.Context, input metatools.ListSkillsInput) (*metatools.ListSkillsOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if h.registry == nil {
		return nil, errors.New("skill registry not configured")
	}

	skills := h.registry.List()
	out := make([]metatools.SkillSummary, 0, len(skills))
	for _, s := range skills {
		if s == nil {
			continue
		}
		out = append(out, metatools.SkillSummary{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			StepCount:   len(s.Steps),
			ToolsetID:   s.ToolsetID,
		})
	}
	return &metatools.ListSkillsOutput{Skills: out}, nil
}

// Describe handles describe_skill.
func (h *SkillsHandler) Describe(ctx context.Context, input metatools.DescribeSkillInput) (*metatools.DescribeSkillOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if h.registry == nil {
		return nil, errors.New("skill registry not configured")
	}

	s, ok := h.registry.Get(input.SkillID)
	if !ok || s == nil {
		return nil, errors.New("skill not found")
	}

	return &metatools.DescribeSkillOutput{
		Skill: skillDefinitionFromInternal(s),
	}, nil
}

// Plan handles plan_skill.
func (h *SkillsHandler) Plan(ctx context.Context, input metatools.PlanSkillInput) (*metatools.PlanSkillOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}

	def, guards, _, err := h.resolveSkill(input.SkillID, input.Skill)
	if err != nil {
		return nil, err
	}

	guards = append(guards, defaultMaxStepsGuard(h.defaults)...)
	plan, err := internalskills.CompilePlan(def, guards)
	if err != nil {
		return nil, err
	}

	return &metatools.PlanSkillOutput{
		Plan: metatools.SkillPlan{
			Name:  plan.Name,
			Steps: stepsToMetatools(plan.Steps),
		},
	}, nil
}

// Run handles run_skill.
func (h *SkillsHandler) Run(ctx context.Context, input metatools.RunSkillInput) (*metatools.RunSkillOutput, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	if err := input.Validate(); err != nil {
		return nil, false, err
	}
	if h.runner == nil {
		return nil, false, errors.New("runner not configured")
	}

	def, guards, skillID, err := h.resolveSkill(input.SkillID, input.Skill)
	if err != nil {
		return skillErrorOutput(err, input.SkillID, false), true, nil
	}

	plan, err := internalskills.CompilePlan(def, append(guards, defaultMaxStepsGuard(h.defaults)...))
	if err != nil {
		return skillErrorOutput(err, skillID, false), true, nil
	}

	effectiveMaxSteps := h.defaults.MaxSteps
	if input.MaxSteps != nil {
		effectiveMaxSteps = *input.MaxSteps
	}
	if effectiveMaxSteps > 0 && len(plan.Steps) > effectiveMaxSteps {
		return skillErrorOutput(skill.ErrMaxStepsExceeded, skillID, false), true, nil
	}

	effectiveMaxCalls := h.defaults.MaxToolCalls
	if input.MaxToolCalls != nil {
		effectiveMaxCalls = *input.MaxToolCalls
	}
	if effectiveMaxCalls > 0 && len(plan.Steps) > effectiveMaxCalls {
		return skillErrorOutput(skill.ErrMaxStepsExceeded, skillID, false), true, nil
	}

	timeout := h.defaults.Timeout
	if input.TimeoutMs != nil {
		timeout = time.Duration(*input.TimeoutMs) * time.Millisecond
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	start := time.Now()
	stepResults, err := skill.Execute(ctx, plan, skillRunner{runner: h.runner})
	duration := int(time.Since(start).Milliseconds())

	output := &metatools.RunSkillOutput{
		Results:    stepResultsToMetatools(stepResults, plan),
		DurationMs: &duration,
	}

	if err != nil {
		output.Error = mapSkillError(err, skillID)
		return output, true, nil
	}
	return output, false, nil
}

type skillRunner struct {
	runner Runner
}

func (r skillRunner) Run(ctx context.Context, step skill.Step) (any, error) {
	result, err := r.runner.Run(ctx, step.ToolID, step.Inputs)
	if err != nil {
		return nil, err
	}
	return result.Structured, nil
}

func (h *SkillsHandler) resolveSkill(id string, def *metatools.SkillDefinition) (skill.Skill, []skill.Guard, string, error) {
	if id != "" {
		if h.registry == nil {
			return skill.Skill{}, nil, "", errors.New("skill registry not configured")
		}
		s, ok := h.registry.Get(id)
		if !ok || s == nil {
			return skill.Skill{}, nil, "", errors.New("skill not found")
		}
		return skill.Skill{Name: s.Name, Steps: s.Steps}, s.Guards, s.ID, nil
	}
	if def == nil {
		return skill.Skill{}, nil, "", errors.New("skill definition required")
	}

	steps := make([]skill.Step, len(def.Steps))
	for i, step := range def.Steps {
		steps[i] = skill.Step{
			ID:     step.ID,
			ToolID: step.ToolID,
			Inputs: step.Inputs,
		}
	}

	if def.ToolsetID != "" && h.toolsets != nil {
		if ts, ok := h.toolsets.Get(def.ToolsetID); ok {
			allowed := make(map[string]struct{})
			for _, id := range ts.ToolIDs() {
				allowed[id] = struct{}{}
			}
			for _, step := range steps {
				if _, ok := allowed[step.ToolID]; !ok {
					return skill.Skill{}, nil, def.Name, skill.ErrToolNotAllowed
				}
			}
		}
	}

	return skill.Skill{Name: def.Name, Steps: steps}, nil, def.Name, nil
}

func skillDefinitionFromInternal(s *internalskills.Skill) metatools.SkillDefinition {
	return metatools.SkillDefinition{
		Name:        s.Name,
		Description: s.Description,
		Steps:       stepsToMetatools(s.Steps),
		ToolsetID:   s.ToolsetID,
	}
}

func stepsToMetatools(steps []skill.Step) []metatools.SkillStep {
	out := make([]metatools.SkillStep, len(steps))
	for i, step := range steps {
		out[i] = metatools.SkillStep{
			ID:     step.ID,
			ToolID: step.ToolID,
			Inputs: step.Inputs,
		}
	}
	return out
}

func stepResultsToMetatools(results []skill.StepResult, plan skill.Plan) []metatools.SkillStepResult {
	stepToolIDs := make(map[string]string, len(plan.Steps))
	for _, step := range plan.Steps {
		stepToolIDs[step.ID] = step.ToolID
	}
	out := make([]metatools.SkillStepResult, len(results))
	for i, res := range results {
		out[i] = metatools.SkillStepResult{
			StepID: res.StepID,
			Value:  res.Value,
		}
		if res.Err != nil {
			out[i].Error = mapSkillError(res.Err, stepToolIDs[res.StepID])
		}
	}
	return out
}

func defaultMaxStepsGuard(defaults SkillDefaults) []skill.Guard {
	if defaults.MaxSteps <= 0 {
		return nil
	}
	return []skill.Guard{skill.MaxStepsGuard(defaults.MaxSteps)}
}

func mapSkillError(err error, toolID string) *metatools.ErrorObject {
	if err == nil {
		return nil
	}
	errObj := merrors.MapToolError(err, toolID, nil, -1)
	if errors.Is(err, skill.ErrInvalidSkillName) ||
		errors.Is(err, skill.ErrInvalidStepID) ||
		errors.Is(err, skill.ErrInvalidToolID) ||
		errors.Is(err, skill.ErrNoSteps) ||
		errors.Is(err, skill.ErrMaxStepsExceeded) ||
		errors.Is(err, skill.ErrToolNotAllowed) {
		errObj.Code = merrors.CodeValidationInput
		errObj.Retryable = false
	}
	return &metatools.ErrorObject{
		Code:        string(errObj.Code),
		Message:     errObj.Message,
		ToolID:      errObj.ToolID,
		Op:          errObj.Op,
		BackendKind: errObj.BackendKind,
		StepIndex:   errObj.StepIndex,
		Retryable:   errObj.Retryable,
		Details:     errObj.Details,
	}
}

func skillErrorOutput(err error, toolID string, includeDuration bool) *metatools.RunSkillOutput {
	out := &metatools.RunSkillOutput{
		Error: mapSkillError(err, toolID),
	}
	if includeDuration {
		zero := 0
		out.DurationMs = &zero
	}
	return out
}
