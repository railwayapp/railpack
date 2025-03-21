package build_llb

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/moby/buildkit/client/llb"
	"github.com/railwayapp/railpack/core/plan"
)

func (g *BuildGraph) GetStateForInput(input plan.Layer) llb.State {
	var state llb.State

	if input.Image != "" {
		state = llb.Image(input.Image, llb.Platform(*g.Platform))
	} else if input.Local {
		state = *g.LocalState
	} else if input.Step != "" {
		if node, exists := g.graph.GetNode(input.Step); exists {
			nodeState := node.(*StepNode).State
			if nodeState == nil {
				return llb.Scratch()
			}
			state = *nodeState
		}
	} else {
		state = llb.Scratch()
	}

	return state
}

func (g *BuildGraph) GetFullStateFromInputs(inputs []plan.Layer) llb.State {
	if len(inputs) == 0 {
		return llb.Scratch()
	}

	fmt.Println("inputs", inputs)

	if len(inputs[0].Include)+len(inputs[0].Exclude) > 0 {
		panic("first input must not have include or exclude paths")
	}

	// Get the base state from the first input
	state := g.GetStateForInput(inputs[0])
	if len(inputs) == 1 {
		return state
	}

	mergeStates := []llb.State{state}
	mergeNames := []string{inputs[0].DisplayName()}

	// Copy from subsequent inputs into the base state
	for _, input := range inputs[1:] {
		inputState := g.GetStateForInput(input)

		// Copy the specified paths (or everything) from this input into our base state
		if len(input.Include) > 0 {
			destState := llb.Scratch()
			destName := input.DisplayName()

			for _, include := range input.Include {
				if input.Local {
					// For local context, always copy into /app
					destPath := filepath.Join("/app", filepath.Base(include))
					destState = destState.File(llb.Copy(inputState, include, destPath, &llb.CopyInfo{
						CopyDirContentsOnly: true,
						CreateDestPath:      true,
						FollowSymlinks:      true,
						AllowWildcard:       true,
						AllowEmptyWildcard:  true,
						ExcludePatterns:     input.Exclude,
					}))
				} else {
					// For other states, handle paths based on whether they're absolute or relative
					srcPath, destPath := resolvePaths(include)

					opts := []llb.ConstraintsOpt{}
					if srcPath == destPath {
						opts = append(opts, llb.WithCustomName(fmt.Sprintf("copy %s", srcPath)))
					}

					destState = destState.File(llb.Copy(inputState, srcPath, destPath, &llb.CopyInfo{
						CopyDirContentsOnly: true,
						CreateDestPath:      true,
						FollowSymlinks:      true,
						AllowWildcard:       true,
						AllowEmptyWildcard:  true,
						ExcludePatterns:     input.Exclude,
					}), opts...)
				}
			}

			mergeStates = append(mergeStates, destState)
			mergeNames = append(mergeNames, destName)
		} else {
			log.Warnf("input %s has no include or exclude paths. This is probably a mistake.", input.Step)
		}
	}

	state = llb.Merge(mergeStates, llb.WithCustomNamef("[railpack] merge %s", strings.Join(mergeNames, ", ")))

	return state
}

func resolvePaths(include string) (srcPath, destPath string) {
	switch {
	case include == "." || include == "/app" || include == "/app/":
		return "/app", "/app"
	case filepath.IsAbs(include):
		return include, include
	default:
		return filepath.Join("/app", include), filepath.Join("/app", include)
	}
}
