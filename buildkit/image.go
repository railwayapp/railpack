package buildkit

import (
	dockerspec "github.com/moby/docker-image-spec/specs-go/v1"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

// Image is the JSON structure which describes some basic information about the image.
// This provides the `application/vnd.oci.image.config.v1+json` mediatype when marshalled to JSON.
type Image struct {
	specs.Image

	// Config defines the execution parameters which should be used as a base when running a container using the image.
	// DockerOCIImageConfig carries the Docker healthcheck field that the OCI config type reserves but does not define.
	Config dockerspec.DockerOCIImageConfig `json:"config"`

	// Variant defines platform variant. To be added to OCI.
	Variant string `json:"variant,omitempty"`
}
