package generate

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/railwayapp/railpack/core/app"
)

type TemplateFileResult struct {
	Filename string
	Contents string
}

// TemplateFiles will look the first file that exists in the list of potential files and render it with the given data
// If no file is found, it will use the default contents and render it with the given data
func (c *GenerateContext) TemplateFiles(potentialFiles []string, defaultContents string, data map[string]any) (*TemplateFileResult, error) {
	filename, contents, err := c.App.ReadFirstFileOf(potentialFiles...)
	if err != nil {
		if !errors.Is(err, app.ErrNoFileFound) {
			return nil, err
		}

		contents = defaultContents
	}

	tmpl, err := template.New(filename).Parse(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return &TemplateFileResult{
		Filename: filename,
		Contents: buf.String(),
	}, nil
}
