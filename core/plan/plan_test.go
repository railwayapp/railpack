package plan

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestSerialization(t *testing.T) {
	jsonPlan := `{
		"steps": [
			{
				"name": "install",
				"commands": [
					{"cmd": "apt-get update"},
					{"cmd": "apt-get install -y curl"}
				],
				"startingImage": "ubuntu:22.04"
			},
			{
				"name": "deps",
				"dependsOn": ["install"],
				"commands": [
					{"path": "/root/.npm", "name": ".npmrc"},
					{"cmd": "npm ci"},
					{"cmd": "npm run build"}
				],
				"useSecrets": true,
				"outputs": [
					"dist",
					"node_modules/.cache"
				],
				"assets": {
					"npmrc": "registry=https://registry.npmjs.org/\n//registry.npmjs.org/:_authToken=${NPM_TOKEN}\nalways-auth=true"
				}
			},
			{
				"name": "build",
				"dependsOn": ["deps"],
				"commands": [
					{"src": ".", "dest": "."},
					{"cmd": "npm run test"},
					{"path": "/usr/local/bin"},
					{"name": "NODE_ENV", "value": "production"}
				],
				"useSecrets": false
			}
		],
		"start": {
			"baseImage": "node:18-slim",
			"cmd": "npm start",
			"paths": ["/usr/local/bin", "/app/node_modules/.bin"]
		},
		"caches": {
			"npm": {
				"directory": "/root/.npm",
				"type": "shared"
			},
			"build-cache": {
				"directory": "node_modules/.cache",
				"type": "locked"
			}
		},
		"secrets": ["NPM_TOKEN", "GITHUB_TOKEN"]
	}`

	var plan1 BuildPlan
	err := json.Unmarshal([]byte(jsonPlan), &plan1)
	require.NoError(t, err)

	serialized, err := json.MarshalIndent(&plan1, "", "  ")
	require.NoError(t, err)

	var plan2 BuildPlan
	err = json.Unmarshal(serialized, &plan2)
	require.NoError(t, err)

	if diff := cmp.Diff(plan1, plan2); diff != "" {
		t.Errorf("plans mismatch (-want +got):\n%s", diff)
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name         string
		plan         *BuildPlan
		expectedPlan *BuildPlan
	}{
		{
			name: "empty plan",
			plan: &BuildPlan{
				Steps:  []Step{},
				Deploy: Deploy{},
			},
			expectedPlan: &BuildPlan{
				Steps:  []Step{},
				Deploy: Deploy{},
			},
		},
		{
			name: "removes empty inputs",
			plan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step1",
						Inputs: []Layer{
							{},
							NewStepLayer("step2"),
						},
					},
				},
				Deploy: Deploy{
					Inputs: []Layer{
						{},
						NewImageLayer("ubuntu:22.04"),
					},
				},
			},
			expectedPlan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step1",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
				},
				Deploy: Deploy{
					Inputs: []Layer{
						NewImageLayer("ubuntu:22.04"),
					},
				},
			},
		},
		{
			name: "removes unreferenced steps",
			plan: &BuildPlan{
				Steps: []Step{
					{
						Name:   "step1",
						Inputs: []Layer{},
					},
					{
						Name:   "step2",
						Inputs: []Layer{},
					},
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name:   "step4",
						Inputs: []Layer{},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step4"),
					Inputs: []Layer{
						NewStepLayer("step3"),
					},
				},
			},
			expectedPlan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name:   "step2",
						Inputs: []Layer{},
					},
					{
						Name:   "step4",
						Inputs: []Layer{},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step4"),
					Inputs: []Layer{
						NewStepLayer("step3"),
					},
				},
			},
		},
		{
			name: "keeps only transitively referenced steps",
			plan: &BuildPlan{
				Steps: []Step{
					{
						Name:   "step1",
						Inputs: []Layer{},
					},
					{
						Name: "step2",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name: "step4",
						Inputs: []Layer{
							NewStepLayer("step3"),
						},
					},
					{
						Name:   "step5",
						Inputs: []Layer{},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step3"),
					Inputs: []Layer{
						NewStepLayer("step5"),
					},
				},
			},
			expectedPlan: &BuildPlan{
				Steps: []Step{
					{
						Name:   "step1",
						Inputs: []Layer{},
					},
					{
						Name: "step2",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name:   "step5",
						Inputs: []Layer{},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step3"),
					Inputs: []Layer{
						NewStepLayer("step5"),
					},
				},
			},
		},
		{
			name: "handles circular dependencies",
			plan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step1",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name: "step2",
						Inputs: []Layer{
							NewStepLayer("step3"),
						},
					},
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
					{
						Name: "step4",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step4"),
				},
			},
			expectedPlan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step4",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
					{
						Name: "step1",
						Inputs: []Layer{
							NewStepLayer("step2"),
						},
					},
					{
						Name: "step2",
						Inputs: []Layer{
							NewStepLayer("step3"),
						},
					},
					{
						Name: "step3",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step4"),
				},
			},
		},
		{
			name: "handles self-referential step",
			plan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step1",
						Inputs: []Layer{
							NewStepLayer("step1"), // References itself
						},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step1"),
				},
			},
			expectedPlan: &BuildPlan{
				Steps: []Step{
					{
						Name: "step1",
						Inputs: []Layer{
							NewStepLayer("step1"),
						},
					},
				},
				Deploy: Deploy{
					Base: NewStepLayer("step1"),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.plan.Normalize()

			// Compare Steps length
			require.Equal(t, len(tc.expectedPlan.Steps), len(tc.plan.Steps), "Steps length mismatch")

			// Create map of expected steps by name for easier comparison
			expectedSteps := make(map[string]Step)
			for _, step := range tc.expectedPlan.Steps {
				expectedSteps[step.Name] = step
			}

			// Verify each step is as expected
			for _, step := range tc.plan.Steps {
				expectedStep, exists := expectedSteps[step.Name]
				require.True(t, exists, "Unexpected step: %s", step.Name)
				require.Equal(t, expectedStep.Inputs, step.Inputs, "Inputs mismatch for step %s", step.Name)
			}

			// Compare Deploy inputs and base
			require.Equal(t, tc.expectedPlan.Deploy.Inputs, tc.plan.Deploy.Inputs, "Deploy inputs mismatch")
			require.Equal(t, tc.expectedPlan.Deploy.Base, tc.plan.Deploy.Base, "Deploy base mismatch")
		})
	}
}

func TestNormalizeWindowsPaths(t *testing.T) {
	plan := &BuildPlan{
		Steps: []Step{
			{
				Name: "build",
				Inputs: []Layer{
					{
						Step: "base",
						Filter: Filter{
							Include: []string{`packages\api`, `src\**`},
							Exclude: []string{`dist\temp`},
						},
					},
				},
				Commands: []Command{
					CopyCommand{Src: `packages\api\package.json`, Dest: `packages\api\package.json`},
					PathCommand{Path: `node_modules\.bin`},
					FileCommand{Path: `config\app`, Name: "config.json"},
				},
			},
		},
		Deploy: Deploy{
			Base:  NewStepLayer("build"),
			Paths: []string{`usr\local\bin`, `app\bin`},
			Inputs: []Layer{
				{
					Step: "build",
					Filter: Filter{
						Include: []string{`out\public`},
					},
				},
			},
		},
		Exclude: []string{`tmp\build`, `cache\mise`},
	}

	plan.Normalize()

	require.Equal(t, []string{"usr/local/bin", "app/bin"}, plan.Deploy.Paths)
	require.Equal(t, []string{"tmp/build", "cache/mise"}, plan.Exclude)
	require.Equal(t, []string{"out/public"}, plan.Deploy.Inputs[0].Include)

	require.Len(t, plan.Steps, 1)
	step := plan.Steps[0]
	require.Equal(t, []string{"packages/api", "src/**"}, step.Inputs[0].Include)
	require.Equal(t, []string{"dist/temp"}, step.Inputs[0].Exclude)

	copyCmd, ok := step.Commands[0].(CopyCommand)
	require.True(t, ok)
	require.Equal(t, "packages/api/package.json", copyCmd.Src)
	require.Equal(t, "packages/api/package.json", copyCmd.Dest)

	pathCmd, ok := step.Commands[1].(PathCommand)
	require.True(t, ok)
	require.Equal(t, "node_modules/.bin", pathCmd.Path)

	fileCmd, ok := step.Commands[2].(FileCommand)
	require.True(t, ok)
	require.Equal(t, "config/app", fileCmd.Path)
}
