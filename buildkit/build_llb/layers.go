package build_llb

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/moby/buildkit/client/llb"
	"github.com/railwayapp/railpack/core/plan"
)

// GetStateForLayer returns the llb.State for a given layer not including any filters (include/exclude)
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

// GetFullStateFromLayers returns the llb.State for a given list of layers including any filters (include/exclude)
// This will attempt to use an llb.Merge operation if possible, otherwise it will use an llb.Copy operation
//
// Merge is more efficient, but if the layers being merged overlap, the the data will be duplicated in the final image resulting in a larger image size
// We try to detect if there are overlaps and fallback to copy everything onto the base state (first layer)
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
		state = copyLayerPaths(state, inputState, input.Filter, input.Local)
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
		destState := copyLayerPaths(llb.Scratch(), inputState, input.Filter, input.Local)
		mergeStates = append(mergeStates, destState)
		mergeNames = append(mergeNames, input.DisplayName())
	}

	return llb.Merge(mergeStates, llb.WithCustomNamef("[railpack] merge %s", strings.Join(mergeNames, ", ")))
}

// copyLayerPaths copies paths from srcState to destState, applying the given filter.
// If isLocal is true, files are copied from local filesystem into /app directory.
// Otherwise paths are copied directly between container locations.
func copyLayerPaths(destState, srcState llb.State, filter plan.Filter, isLocal bool) llb.State {
	for _, include := range filter.Include {
		srcPath, destPath := resolvePaths(include, isLocal)

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
			ExcludePatterns:     filter.Exclude,
		}), opts...)
	}
	return destState
}

// shouldLLBMerge determines if a set of layers should be merged based on path overlaps.
// We should not merge layers if:
// - The non-first layer has no include filters
// - Any layer includes the root path "/"
// - Any layer pulls from a local filesystem
// - Any layer has overlapping paths with subsequent layers
func shouldLLBMerge(layers []plan.Layer) bool {
	for i, layer := range layers {
		if i != 0 && layer.Include == nil {
			return false
		}

		if slices.Contains(layer.Include, "/") {
			return false
		}

		if layer.Local {
			return false
		}

		for j := i + 1; j < len(layers); j++ {
			if hasPathOverlap(layer.Include, layers[j].Include) {
				return false
			}
		}
	}
	return true
}

// hasPathOverlap checks if two slices of paths have any overlapping paths.
// Paths overlap if they are identical or if one is a subdirectory of the other.
// For example:
//
//	hasPathOverlap([]string{"/app/dist"}, []string{"/app"}) // returns true
//	hasPathOverlap([]string{"/app-foo"}, []string{"/app"}) // returns false
func hasPathOverlap(paths1, paths2 []string) bool {
	for _, p1 := range paths1 {
		p1Clean := path.Clean(p1)
		if !strings.HasSuffix(p1Clean, "/") {
			p1Clean = p1Clean + "/"
		}

		for _, p2 := range paths2 {
			p2Clean := path.Clean(p2)
			if !strings.HasSuffix(p2Clean, "/") {
				p2Clean = p2Clean + "/"
			}

			// Check direct path match or if one is a subdirectory of the other
			if p1Clean == p2Clean || strings.HasPrefix(p1Clean, p2Clean) || strings.HasPrefix(p2Clean, p1Clean) {
				return true
			}
		}
	}
	return false
}

// resolvePaths determines source and destination paths based on the include path and whether it's local.
// For local paths, only the basename is preserved when copying to /app directory.
// For container paths, the full relative path structure is preserved under /app.
func resolvePaths(include string, isLocal bool) (srcPath, destPath string) {
	if isLocal {
		// convert a local path reference to fully qualified container path
		return include, filepath.Join("/app", filepath.Base(include))
	}

	switch {
	case include == "." || include == "/app" || include == "/app/":
		return "/app", "/app"
	case filepath.IsAbs(include):
		return include, include
	default:
		return filepath.Join("/app", include), filepath.Join("/app", include)
	}
}
