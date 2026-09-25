package plan

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name            string
		command         Command
		expectedJSON    string
		unmarshalString string
	}{
		// Exec
		{
			name:            "exec command without custom name",
			command:         NewExecShellCommand("echo hello", ExecOptions{CustomName: "echo hello"}),
			expectedJSON:    `{"cmd":"sh -c 'echo hello'","customName":"echo hello"}`,
			unmarshalString: "echo hello",
		},
		{
			name:            "exec command with custom name",
			command:         NewExecShellCommand("echo hello", ExecOptions{CustomName: "Say Hello"}),
			expectedJSON:    `{"cmd":"sh -c 'echo hello'","customName":"Say Hello"}`,
			unmarshalString: "RUN#Say Hello:echo hello",
		},

		// Path
		{
			name:            "path command",
			command:         NewPathCommand("/usr/local/bin"),
			expectedJSON:    `{"path":"/usr/local/bin"}`,
			unmarshalString: "PATH:/usr/local/bin",
		},

		// Copy
		{
			name:            "copy command",
			command:         NewCopyCommand("src.txt", "dst.txt"),
			expectedJSON:    `{"src":"src.txt","dest":"dst.txt"}`,
			unmarshalString: "COPY:src.txt dst.txt",
		},

		// File
		{
			name:            "file command without custom name",
			command:         NewFileCommand("/etc/conf", "config.yaml"),
			expectedJSON:    `{"path":"/etc/conf","name":"config.yaml"}`,
			unmarshalString: "FILE:/etc/conf config.yaml",
		},
		{
			name:            "file command with custom name",
			command:         NewFileCommand("/etc/conf", "config.yaml", FileOptions{CustomName: "Config File"}),
			expectedJSON:    `{"path":"/etc/conf","name":"config.yaml","customName":"Config File"}`,
			unmarshalString: "FILE#Config File:/etc/conf config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshalling to JSON object
			data, err := json.Marshal(tt.command)
			require.NoError(t, err, "failed to marshal command")
			require.Equal(t, string(data), tt.expectedJSON, "marshal result")

			// Test unmarshalling from JSON object
			cmd, err := UnmarshalCommand([]byte(tt.expectedJSON))
			require.NoError(t, err, "failed to unmarshal JSON command")

			// Marshal again to verify it produces the same result
			roundTrip, err := json.Marshal(cmd)
			require.NoError(t, err, "failed to marshal unmarshalled command")
			require.Equal(t, string(roundTrip), tt.expectedJSON, "round-trip JSON result")

			// Test unmarshalling from string format
			if tt.unmarshalString != "" {
				cmd, err = UnmarshalCommand([]byte(tt.unmarshalString))
				require.NoError(t, err, "failed to unmarshal string command")

				// Marshal to JSON to verify it produces the same object
				roundTrip, err = json.Marshal(cmd)
				require.NoError(t, err, "failed to marshal string-unmarshalled command")
				require.Equal(t, string(roundTrip), tt.expectedJSON, "string unmarshal to JSON result")

			}
		})
	}
}
func TestUnmarshalCommandFromJSONRawMessage(t *testing.T) {
	tests := []struct {
		name      string
		jsonRaw   string
		checkType func(Command) bool
	}{
		{
			name:    "shorthand COPY with JSON quotes",
			jsonRaw: `"COPY:src.txt dst.txt"`,
			checkType: func(c Command) bool {
				copyCmd, ok := c.(CopyCommand)
				return ok && copyCmd.Src == "src.txt" && copyCmd.Dest == "dst.txt"
			},
		},
		{
			name:    "shorthand PATH with JSON quotes",
			jsonRaw: `"PATH:/usr/local/bin"`,
			checkType: func(c Command) bool {
				pathCmd, ok := c.(PathCommand)
				return ok && pathCmd.Path == "/usr/local/bin"
			},
		},
		{
			name:    "shorthand RUN with custom name and JSON quotes",
			jsonRaw: `"RUN#Say Hello:echo hello"`,
			checkType: func(c Command) bool {
				execCmd, ok := c.(ExecCommand)
				return ok && execCmd.CustomName == "Say Hello" && execCmd.Cmd == "sh -c 'echo hello'"
			},
		},
		{
			name:    "shorthand FILE with JSON quotes",
			jsonRaw: `"FILE:/etc/conf config.yaml"`,
			checkType: func(c Command) bool {
				fileCmd, ok := c.(FileCommand)
				return ok && fileCmd.Path == "/etc/conf" && fileCmd.Name == "config.yaml"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := UnmarshalCommand([]byte(tt.jsonRaw))
			require.NoError(t, err)
			require.True(t, tt.checkType(cmd), "Command was not parsed into expected type: %+v", cmd)
		})
	}
}
