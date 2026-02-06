package skills

import "github.com/jonwraymond/toolcompose/skill"

// CompilePlan validates guards and builds a deterministic plan.
func CompilePlan(def skill.Skill, guards []skill.Guard) (skill.Plan, error) {
	for _, guard := range guards {
		if guard == nil {
			continue
		}
		if err := guard.Validate(def); err != nil {
			return skill.Plan{}, err
		}
	}
	planner := skill.NewPlanner()
	return planner.Plan(def)
}
