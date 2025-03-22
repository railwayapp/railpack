package plan

const (
	RailpackBuilderImage = "ghcr.io/railwayapp/railpack-builder:latest"
	RailpackRuntimeImage = "ghcr.io/railwayapp/railpack-runtime:latest"
)

type BuildPlan struct {
	Steps   []Step            `json:"steps,omitempty"`
	Caches  map[string]*Cache `json:"caches,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Deploy  Deploy            `json:"deploy,omitempty"`
}

type Deploy struct {
	// The base layer for the deploy step
	Base Layer `json:"base,omitempty"`

	// The layers for the deploy step
	Inputs []Layer `json:"inputs,omitempty"`

	// The command to run in the container
	StartCmd string `json:"startCommand,omitempty"`

	// The variables available to this step. The key is the name of the variable that is referenced in a variable command
	Variables map[string]string `json:"variables,omitempty"`

	// The paths to prepend to the $PATH environment variable
	Paths []string `json:"paths,omitempty"`
}

func NewBuildPlan() *BuildPlan {
	return &BuildPlan{
		Steps:   []Step{},
		Deploy:  Deploy{},
		Caches:  make(map[string]*Cache),
		Secrets: []string{},
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}

func (p *BuildPlan) Normalize() {
	// Remove empty inputs from steps
	for _, step := range p.Steps {
		normalizedInputs := []Layer{}
		for _, input := range step.Inputs {
			if !input.IsEmpty() {
				normalizedInputs = append(normalizedInputs, input)
			}
		}
		step.Inputs = normalizedInputs
	}

	// Remove empty inputs from deploy
	normalizedDeployInputs := []Layer{}
	for _, input := range p.Deploy.Inputs {
		if !input.IsEmpty() {
			normalizedDeployInputs = append(normalizedDeployInputs, input)
		}
	}
	p.Deploy.Inputs = normalizedDeployInputs

	// Track which steps are referenced by other steps or deploy
	referencedSteps := make(map[string]bool)

	// Check Deploy.Base for step references
	if p.Deploy.Base.Step != "" {
		referencedSteps[p.Deploy.Base.Step] = true
	}

	for _, input := range p.Deploy.Inputs {
		if input.Step != "" {
			referencedSteps[input.Step] = true
		}
	}

	for _, step := range p.Steps {
		for _, input := range step.Inputs {
			if input.Step != "" {
				referencedSteps[input.Step] = true
			}
		}
	}

	// Keep only steps that are referenced (or all if none are referenced)
	normalizedSteps := []Step{}
	for _, step := range p.Steps {
		if referencedSteps[step.Name] || len(referencedSteps) == 0 {
			normalizedSteps = append(normalizedSteps, step)
		}
	}
	p.Steps = normalizedSteps
}
