package plan

import (
	"fmt"
	"path/filepath"

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
	// Normalize paths in deploy and steps to use forward slashes
	for i, path := range p.Deploy.Paths {
		p.Deploy.Paths[i] = filepath.ToSlash(path)
	}
	for i, excl := range p.Exclude {
		p.Exclude[i] = filepath.ToSlash(excl)
	}

	for i := range p.Steps {
		for j, cmd := range p.Steps[i].Commands {
			switch c := cmd.(type) {
			case CopyCommand:
				c.Src = filepath.ToSlash(c.Src)
				c.Dest = filepath.ToSlash(c.Dest)
				p.Steps[i].Commands[j] = c
			case PathCommand:
				c.Path = filepath.ToSlash(c.Path)
				p.Steps[i].Commands[j] = c
			case FileCommand:
				c.Path = filepath.ToSlash(c.Path)
				p.Steps[i].Commands[j] = c
			}
		}
		for j := range p.Steps[i].Inputs {
			p.Steps[i].Inputs[j].Include = normalizeFilterPaths(p.Steps[i].Inputs[j].Include)
			p.Steps[i].Inputs[j].Exclude = normalizeFilterPaths(p.Steps[i].Inputs[j].Exclude)
		}
	}

	if p.Deploy.Inputs != nil {
		for j := range p.Deploy.Inputs {
			p.Deploy.Inputs[j].Include = normalizeFilterPaths(p.Deploy.Inputs[j].Include)
			p.Deploy.Inputs[j].Exclude = normalizeFilterPaths(p.Deploy.Inputs[j].Exclude)
		}
	}

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
	// We use maxIterations to prevent infinite loops from circular dependencies.
	// If we exceed this maximum possible number of unique edges, we can safely
	// break as we've already collected all reachable steps.
	maxIterations := len(p.Steps) * len(p.Steps) // Maximum possible unique edges in a directed graph

	for range maxIterations {
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
