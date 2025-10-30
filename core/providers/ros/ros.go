package ros

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const ROS_DEFAULT_IMAGE = "osrf/ros:kilted-desktop"

type RosProvider struct{}

func (p *RosProvider) Name() string {
	return "ros"
}

func (p *RosProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasFile("package.xml"), nil
}

func (p *RosProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *RosProvider) StartCommandHelp() string {
	return ""
}

func (p *RosProvider) Plan(ctx *generate.GenerateContext) error {
	baseImage := ctx.NewImageStep("image", func(options *generate.BuildStepOptions) string {
		return ROS_DEFAULT_IMAGE
	})

	packages := ctx.NewCommandStep("apt")
	packages.AddCommands([]plan.Command{
		plan.NewExecCommand("apt-get update"),
		plan.NewExecCommand("apt-get install -y python3-colcon-common-extensions python3-rosdep"),
	})
	packages.AddInput(plan.NewStepLayer(baseImage.Name()))

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("rosdep update"),
		plan.NewExecCommand("rosdep install --from-paths src --ignore-src -r -y"),
		plan.NewExecCommand("colcon build"),
	})

	launchFiles, err := ctx.App.FindFiles("launch/*.{xml,py}")

	ctx.Deploy.StartCmd = "source install/setup.bash"
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(build.Name()),
	})
	ctx.Deploy.Base.Image = ROS_DEFAULT_IMAGE
	if err == nil && len(launchFiles) > 0 {
		ctx.Deploy.StartCmd += fmt.Sprintf(" && ros launch %s", launchFiles[0])
	}

	return nil
}

func (p *RosProvider) installErlang(step *generate.MiseStepBuilder) {
	step.Default("erlang", "latest")
}
