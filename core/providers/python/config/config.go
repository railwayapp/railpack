package config

type PythonConfig struct {
	Version       string `json:"version,omitempty" jsonschema:"description=Override the Python version for the python provider"`
	DjangoAppName string `json:"djangoAppName,omitempty" jsonschema:"description=Specify the Django WSGI application module for the python provider"`
}
