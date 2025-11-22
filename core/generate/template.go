package generate

import (
	"bytes"
	"fmt"
	"text/template"
)

type TemplateFileResult struct {
	Filename string
	Contents string
}

// TemplateFiles will look the first file that exists in the list of potential files and render it with the given data
// If no file is found, it will use the default contents and render it with the given data
func (c *GenerateContext) TemplateFiles(potentialFiles []string, defaultContents string, data map[string]interface{}) (*TemplateFileResult, error) {
	contents := defaultContents
	filename := ""

	for _, potentialFilename := range potentialFiles {
		if c.App.HasFile(potentialFilename) {
			c, err := c.App.ReadFile(potentialFilename)
			if err != nil {
				return nil, err
			}

			contents = c
			filename = potentialFilename

			break
		}
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
