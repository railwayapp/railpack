package build_llb

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/moby/buildkit/client/llb"
	"github.com/railwayapp/railpack/core/plan"
)

func (g *BuildGraph) GetStateForLayer(layer plan.Layer) llb.State {
	var state llb.State

	if layer.Image != "" {
		state = llb.Image(layer.Image, llb.Platform(*g.Platform))
	} else if layer.Local {
		state = *g.LocalState
	} else if layer.Step != "" {
		if node, exists := g.graph.GetNode(layer.Step); exists {
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

func (g *BuildGraph) GetFullStateFromLayers(layers []plan.Layer) llb.State {
	if len(layers) == 0 {
		return llb.Scratch()
	}

	if len(layers[0].Include)+len(layers[0].Exclude) > 0 {
		panic("first input must not have include or exclude paths")
	}

	// Get the base state from the first input
	state := g.GetStateForLayer(layers[0])
	if len(layers) == 1 {
		return state
	}

	shouldMerge := shouldLLBMerge(layers)
	if shouldMerge {
		return g.getMergeState(layers)
	}

	return g.getCopyState(layers)
}

func (g *BuildGraph) getCopyState(layers []plan.Layer) llb.State {
	state := g.GetStateForLayer(layers[0])
	if len(layers) == 1 {
		return state
	}

	for _, input := range layers[1:] {
		inputState := g.GetStateForLayer(input)
		state = copyLayerPaths(state, inputState, input)
	}
	return state
}

func (g *BuildGraph) getMergeState(layers []plan.Layer) llb.State {
	mergeStates := []llb.State{g.GetStateForLayer(layers[0])}
	mergeNames := []string{layers[0].DisplayName()}

	for _, input := range layers[1:] {
		if len(input.Include) == 0 {
			log.Warnf("input %s has no include or exclude paths. This is probably a mistake.", input.Step)
		}
		inputState := g.GetStateForLayer(input)
		destState := copyLayerPaths(llb.Scratch(), inputState, input)
		mergeStates = append(mergeStates, destState)
		mergeNames = append(mergeNames, input.DisplayName())
	}

	return llb.Merge(mergeStates, llb.WithCustomNamef("[railpack] merge %s", strings.Join(mergeNames, ", ")))
}

func copyLayerPaths(destState, srcState llb.State, layer plan.Layer) llb.State {
	for _, include := range layer.Include {
		var srcPath, destPath string
		if layer.Local {
			srcPath = include
			destPath = filepath.Join("/app", filepath.Base(include))
		} else {
			srcPath, destPath = resolvePaths(include)
		}

		opts := []llb.ConstraintsOpt{}
		if srcPath == destPath {
			opts = append(opts, llb.WithCustomName(fmt.Sprintf("copy %s", srcPath)))
		}

		destState = destState.File(llb.Copy(srcState, srcPath, destPath, &llb.CopyInfo{
			CopyDirContentsOnly: true,
			CreateDestPath:      true,
			FollowSymlinks:      true,
			AllowWildcard:       true,
			AllowEmptyWildcard:  true,
			ExcludePatterns:     layer.Exclude,
		}), opts...)
	}
	return destState
}

func shouldLLBMerge(layers []plan.Layer) bool {
	for i, layer := range layers {
		for j := i + 1; j < len(layers); j++ {
			if hasPathOverlap(layer.Include, layers[j].Include) {
				return false
			}
		}
	}
	return true
}

func hasPathOverlap(paths1, paths2 []string) bool {
	for _, p1 := range paths1 {
		if slices.Contains(paths2, p1) {
			return true
		}
	}
	return false
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
