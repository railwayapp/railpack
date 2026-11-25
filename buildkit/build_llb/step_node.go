package build_llb

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/railwayapp/railpack/core/plan"
)

type StepNode struct {
	Step       *plan.Step
	State      *llb.State
	parents    []*StepNode
	children   []*StepNode
	Processed  bool
	InProgress bool

	InputEnv  BuildEnvironment
	OutputEnv BuildEnvironment
}

func (node *StepNode) getPathList() []string {
	pathList := make([]string, 0)
	pathList = append(pathList, node.OutputEnv.PathList...)
	return pathList
}
