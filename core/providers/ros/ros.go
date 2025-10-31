package ros

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

const (
	ROS_IMAGE           = "osrf/ros:%s-desktop"
	ROS_VERSION_DEFAULT = "kilted"
)

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
	rosVersion := ctx.Env.GetVariable(ctx.Env.ConfigVariable("ROS_VERSION"))
	if rosVersion == "" {
		rosVersion = ROS_VERSION_DEFAULT
	}
	baseImageName := fmt.Sprintf(ROS_IMAGE, rosVersion)
	baseImage := ctx.NewImageStep("image", func(options *generate.BuildStepOptions) string {
		return baseImageName
	})

	packages := ctx.NewCommandStep("apt")
	packages.AddCommands([]plan.Command{
		plan.NewExecCommand("apt-get update"),
		plan.NewExecCommand(fmt.Sprintf("apt-get install -y python3-colcon-common-extensions python3-rosdep ros-dev-tools ros-%s-desktop", rosVersion)),
	})
	if ctx.App.HasMatch("**/CMakeLists.txt") {
		packages.AddCommand(plan.NewExecCommand(fmt.Sprintf("apt-get install -y ros-%s-ament-cmake ros-%s-ament-cmake-core ros-%s-ament-cmake-python", rosVersion, rosVersion, rosVersion)))
	}
	packages.AddInput(plan.NewStepLayer(baseImage.Name()))

	build := ctx.NewCommandStep("build")
	build.AddInput(plan.NewStepLayer(packages.Name()))
	build.AddInput(ctx.NewLocalLayer())
	build.AddCommands([]plan.Command{
		plan.NewExecCommand("rosdep update"),
		plan.NewExecCommand("rosdep install --from-paths src --ignore-src -r -y"),
		plan.NewExecCommand(fmt.Sprintf("bash -c 'source /opt/ros/%s/setup.bash && colcon build'", rosVersion)),
	})

	launchFile := ctx.Env.GetVariable(ctx.Env.ConfigVariable("ROS_LAUNCH_FILE"))
	if launchFile == "" {
		launchFiles, err := ctx.App.FindFiles("launch/*.{xml,py}")
		if err == nil && len(launchFiles) > 0 {
			launchFile = launchFiles[0]
		}
	} else {
		launchFile = "launch/" + launchFile
	}

	ctx.Deploy.StartCmd = fmt.Sprintf("source /opt/ros/%s/setup.bash && source install/setup.bash", rosVersion)
	ctx.Deploy.AddInputs([]plan.Layer{
		plan.NewStepLayer(packages.Name()),
		plan.NewStepLayer(build.Name(), plan.NewIncludeFilter([]string{"."})),
	})
	ctx.Deploy.Base.Image = baseImageName
	if launchFile != "" {
		ctx.Deploy.StartCmd += fmt.Sprintf(" && ros2 launch %s", launchFile)
	}

	return nil
}
