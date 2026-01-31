package plan

const (
	RailpackBuilderImage = "ghcr.io/railwayapp/railpack-builder:latest"
	RailpackRuntimeImage = "ghcr.io/railwayapp/railpack-runtime:latest"
)

// serialized to railpack.json
type BuildPlan struct {
	Steps   []Step            `json:"steps,omitempty"`
	Caches  map[string]*Cache `json:"caches,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Deploy  Deploy            `json:"deploy,omitempty"`
	Exclude []string          `json:"exclude,omitempty"`
	Include []string          `json:"include,omitempty"`
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
		Exclude: []string{},
		Include: []string{},
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}

func (p *BuildPlan) Normalize() {
	// Remove empty inputs from steps
	for i := range p.Steps {
		if p.Steps[i].Inputs == nil {
			continue
		}
		normalizedInputs := []Layer{}
		for _, input := range p.Steps[i].Inputs {
			if !input.IsEmpty() {
				normalizedInputs = append(normalizedInputs, input)
			}
		}
		p.Steps[i].Inputs = normalizedInputs
	}

	// Remove empty inputs from deploy
	if p.Deploy.Inputs != nil {
		normalizedDeployInputs := []Layer{}
		for _, input := range p.Deploy.Inputs {
			if !input.IsEmpty() {
				normalizedDeployInputs = append(normalizedDeployInputs, input)
			}
		}
		if len(normalizedDeployInputs) == 0 {
			p.Deploy.Inputs = nil
		} else {
			p.Deploy.Inputs = normalizedDeployInputs
		}
	}

	// Track which steps are referenced by deploy or transitively referenced steps
	referencedSteps := make(map[string]bool)

	// Start with steps referenced directly by deploy
	if p.Deploy.Base.Step != "" {
		referencedSteps[p.Deploy.Base.Step] = true
	}

	if p.Deploy.Inputs != nil {
		for _, input := range p.Deploy.Inputs {
			if input.Step != "" {
				referencedSteps[input.Step] = true
			}
		}
	}

	// Keep finding new referenced steps until no more are found
	// Use a map to track which steps we've already checked to avoid infinite loops
	checkedSteps := make(map[string]bool)
	maxIterations := len(p.Steps) * len(p.Steps) // Maximum possible unique edges in a directed graph
	iterations := 0

	for {
		if iterations >= maxIterations {
			// We've exceeded the maximum possible number of unique edges
			// This means we have a circular dependency, but we've already
			// collected all reachable steps, so we can break
			break
		}
		iterations++

		newReferences := false
		for _, step := range p.Steps {
			// Skip if this step isn't referenced
			if !referencedSteps[step.Name] {
				continue
			}

			// Skip if we've already checked this step's inputs
			if checkedSteps[step.Name] {
				continue
			}

			// Mark this step as checked
			checkedSteps[step.Name] = true

			// Check this step's inputs for references
			if step.Inputs != nil {
				for _, input := range step.Inputs {
					if input.Step != "" && !referencedSteps[input.Step] {
						referencedSteps[input.Step] = true
						newReferences = true
					}
				}
			}
		}
		if !newReferences {
			break
		}
	}

	// Keep only steps that are referenced
	if len(referencedSteps) > 0 {
		normalizedSteps := make([]Step, 0, len(p.Steps))
		for _, step := range p.Steps {
			if referencedSteps[step.Name] {
				normalizedSteps = append(normalizedSteps, step)
			}
		}
		p.Steps = normalizedSteps
	}
}
