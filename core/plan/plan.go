package plan

import (
	"fmt"

	"github.com/railwayapp/railpack/core/mise"
)

var (
	RailpackRuntimeImage = fmt.Sprintf("ghcr.io/railwayapp/railpack-runtime:mise-%s", mise.Version)
)

// serialized to railpack.json
type BuildPlan struct {
	Steps   []Step            `json:"steps,omitempty"`
	Caches  map[string]*Cache `json:"caches,omitempty"`
	Secrets []string          `json:"secrets,omitempty"`
	Deploy  Deploy            `json:"deploy"`
	Exclude []string          `json:"exclude,omitempty"`
}

type Deploy struct {
	// The base layer for the deploy step
	Base Layer `json:"base"`

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
	}
}

func (p *BuildPlan) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}

func (p *BuildPlan) Normalize() {
	// Remove empty inputs from steps
	for i := range p.Steps {
		var normalizedInputs []Layer
		for _, input := range p.Steps[i].Inputs {
			if !input.IsEmpty() {
				normalizedInputs = append(normalizedInputs, input)
			}
		}
		p.Steps[i].Inputs = normalizedInputs
	}

	// Remove empty inputs from deploy
	if p.Deploy.Inputs != nil {
		var normalizedDeployInputs []Layer
		for _, input := range p.Deploy.Inputs {
			if !input.IsEmpty() {
				normalizedDeployInputs = append(normalizedDeployInputs, input)
			}
		}
		p.Deploy.Inputs = normalizedDeployInputs
	}

	referencedSteps := p.referencedSteps()
	if len(referencedSteps) == 0 {
		// Plans without a deploy step reference keep all build steps.
		return
	}

	normalizedSteps := make([]Step, 0, len(referencedSteps))
	for _, step := range p.Steps {
		if referencedSteps[step.Name] {
			normalizedSteps = append(normalizedSteps, step)
		}
	}
	p.Steps = normalizedSteps
}

// Walks upstream from the step references in deploy.
func (p *BuildPlan) referencedSteps() map[string]bool {
	referenced := make(map[string]bool)
	var pending []string
	push := func(name string) {
		if name == "" || referenced[name] {
			return
		}
		referenced[name] = true
		pending = append(pending, name)
	}

	push(p.Deploy.Base.Step)
	for _, input := range p.Deploy.Inputs {
		push(input.Step)
	}
	if len(pending) == 0 {
		return referenced
	}

	stepsByName := make(map[string]*Step, len(p.Steps))
	for i := range p.Steps {
		step := &p.Steps[i]
		stepsByName[step.Name] = step
	}

	for len(pending) > 0 {
		name := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		step := stepsByName[name]
		if step == nil {
			continue
		}
		for _, input := range step.Inputs {
			push(input.Step)
		}
	}
	return referenced
}
