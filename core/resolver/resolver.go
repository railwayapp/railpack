package resolver

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/railwayapp/railpack/core/mise"
)

const (
	DefaultSource = "railpack default"
)

type Resolver struct {
	mise             *mise.Mise
	packages         map[string]*RequestedPackage
	previousVersions map[string]string
}

type RequestedPackage struct {
	Name    string
	Version string
	Source  string

	// If set, the resolver will use this function to determine if a version is available
	IsVersionAvailable func(version string) bool

	// If true, the package will not be installed with mise
	// but the version will still be resolved and displayed in the output
	SkipMiseInstall bool
}

type ResolvedPackage struct {
	Name             string  `json:"name"`
	RequestedVersion *string `json:"requestedVersion,omitempty"`
	ResolvedVersion  *string `json:"resolvedVersion,omitempty"`
	Source           string  `json:"source"`
}

type PackageRef struct {
	Name string
}

func NewRequestedPackage(name, defaultVersion string) *RequestedPackage {
	return &RequestedPackage{
		Name:    name,
		Version: defaultVersion,
		Source:  DefaultSource,
	}
}

func (p *RequestedPackage) SetVersion(version, source string) *RequestedPackage {
	p.Version = version
	p.Source = source
	return p
}

func NewResolver(miseDir string) (*Resolver, error) {
	mise, err := mise.New(miseDir)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		mise:             mise,
		packages:         make(map[string]*RequestedPackage),
		previousVersions: make(map[string]string),
	}, nil
}

func (r *Resolver) ResolvePackages() (map[string]*ResolvedPackage, error) {
	resolvedPackages := make(map[string]*ResolvedPackage)

	for name, pkg := range r.packages {
		fuzzyVersion := resolveToFuzzyVersion(pkg.Version)

		var latestVersion string

		// If there is a custom version validator, we get possible versions and pick the latest one that matches
		if pkg.IsVersionAvailable != nil {
			versions, err := r.mise.GetAllVersions(name, fuzzyVersion)
			if err != nil {
				return nil, err
			}

			for i := len(versions) - 1; i >= 0; i-- {
				if pkg.IsVersionAvailable(versions[i]) {
					latestVersion = versions[i]
					break
				}
			}

			if latestVersion == "" {
				return nil, fmt.Errorf("no version available for %s %s", name, pkg.Version)
			}
		} else {
			// Otherwise, we just get the latest version
			var err error
			latestVersion, err = r.mise.GetLatestVersion(name, fuzzyVersion)
			if err != nil {
				return nil, err
			}
		}

		log.Debugf("Resolved package version %s %s to %s from %s", name, pkg.Version, latestVersion, pkg.Source)

		resolvedPkg := &ResolvedPackage{
			Name:             name,
			RequestedVersion: &pkg.Version,
			ResolvedVersion:  &latestVersion,
			Source:           pkg.Source,
		}

		resolvedPackages[name] = resolvedPkg
	}

	return resolvedPackages, nil
}

func (r *Resolver) Get(name string) *RequestedPackage {
	return r.packages[name]
}

func (r *Resolver) Default(name, defaultVersion string) PackageRef {
	r.packages[name] = NewRequestedPackage(name, defaultVersion)

	// If there is a previous version of the package, use that instead of the default version
	if r.previousVersions[name] != "" && r.previousVersions[name] != defaultVersion {
		r.Version(PackageRef{Name: name}, r.previousVersions[name], "previous installed version")
	}

	return PackageRef{Name: name}
}

func (r *Resolver) Version(ref PackageRef, version, source string) PackageRef {
	if pkg, exists := r.packages[ref.Name]; exists {
		pkg.SetVersion(strings.TrimSpace(version), source)
	}
	return ref
}

func (r *Resolver) SetPreviousVersion(name, version string) {
	r.previousVersions[name] = version
}

func (r *Resolver) SetVersionAvailable(ref PackageRef, isVersionAvailable func(version string) bool) {
	r.packages[ref.Name].IsVersionAvailable = isVersionAvailable
}

func (r *Resolver) SetSkipMiseInstall(ref PackageRef, skipMiseInstall bool) {
	r.packages[ref.Name].SkipMiseInstall = skipMiseInstall
}
