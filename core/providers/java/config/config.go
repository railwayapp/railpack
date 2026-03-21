package config

type JavaConfig struct {
	Version       string `json:"version,omitempty" jsonschema:"description=Override the JDK version for the java provider"`
	GradleVersion string `json:"gradleVersion,omitempty" jsonschema:"description=Override the Gradle version for the java provider"`
}
