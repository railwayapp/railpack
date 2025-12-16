package node

import (
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

// RSC (React Server Components) vulnerability information
// CVE-2025-55182: Critical RCE vulnerability (CVSS 10.0)
// CVE-2025-55183: Source Code Exposure (CVSS 5.3)
// CVE-2025-55184: Denial of Service (CVSS 7.5)
// CVE-2025-67779: Denial of Service (CVSS 7.5)
// See: https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components
// See: https://react.dev/blog/2025/12/11/denial-of-service-and-source-code-exposure-in-react-server-components

// Vulnerable React RSC packages that need version checking
var vulnerableRSCPackages = []string{
	"react-server-dom-webpack",
	"react-server-dom-parcel",
	"react-server-dom-turbopack",
}

// Frameworks/bundlers that use RSC and may have vulnerable dependencies
// These don't have specific version fixes - instead we upgrade their RSC dependencies
var rscAffectedFrameworks = []string{
	"waku",
	"@parcel/rsc",
	"@vitejs/plugin-rsc",
	"rwsdk",
}

// Fixed versions for React RSC packages by major.minor release line
var rscFixedVersions = map[string]string{
	"19.0": "19.0.3",
	"19.1": "19.1.4",
	"19.2": "19.2.3",
}

// Fixed versions for Next.js by major.minor release line
// See: https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components#update-next-js
var nextjsFixedVersions = map[string]string{
	"13.3": "14.2.35",
	"13.4": "14.2.35",
	"13.5": "14.2.35",
	"14.0": "14.2.35",
	"14.1": "14.2.35",
	"14.2": "14.2.35",
	"15.0": "15.0.7",
	"15.1": "15.1.11",
	"15.2": "15.2.8",
	"15.3": "15.3.8",
	"15.4": "15.4.10",
	"15.5": "15.5.9",
	"16.0": "16.0.10",
}

// React Router versions that include vulnerable RSC
// React Router with unstable RSC APIs needs react-server-dom-* packages upgraded
var reactRouterFixedVersions = map[string]string{
	"7.0": "7.0.3",
	"7.1": "7.1.6",
	"7.2": "7.2.3",
	"7.3": "7.3.3",
	"7.4": "7.4.2",
	"7.5": "7.5.3",
	"7.6": "7.6.3",
}

// VulnerabilityType categorizes the type of RSC vulnerability
type VulnerabilityType int

const (
	VulnTypeRSCPackage   VulnerabilityType = iota // Direct RSC package (react-server-dom-*)
	VulnTypeNextJS                                // Next.js framework
	VulnTypeReactRouter                           // React Router with RSC
	VulnTypeRSCFramework                          // Other RSC frameworks (waku, @parcel/rsc, etc.)
)

// SecurityVulnerabilityInfo contains information about a detected security vulnerability
type SecurityVulnerabilityInfo struct {
	Package        string
	CurrentVersion string
	FixedVersion   string
	Location       string            // e.g., "root" or workspace package path
	Type           VulnerabilityType // Type of vulnerability
}

// checkSecurityVulnerabilities checks for vulnerable packages in the workspace
func (p *NodeProvider) checkSecurityVulnerabilities(ctx *generate.GenerateContext) []SecurityVulnerabilityInfo {
	var vulnerabilities []SecurityVulnerabilityInfo

	if p.workspace == nil {
		return vulnerabilities
	}

	vulnerabilities = append(vulnerabilities, p.checkPackageVulnerabilities(p.workspace.Root.PackageJson, "root")...)

	for _, workspacePkg := range p.workspace.Packages {
		vulnerabilities = append(vulnerabilities, p.checkPackageVulnerabilities(workspacePkg.PackageJson, workspacePkg.Path)...)
	}

	return vulnerabilities
}

// checkPackageVulnerabilities checks a single package.json for RSC vulnerabilities
func (p *NodeProvider) checkPackageVulnerabilities(pkgJson *PackageJson, location string) []SecurityVulnerabilityInfo {
	var vulnerabilities []SecurityVulnerabilityInfo

	// Check for direct RSC package vulnerabilities
	for _, pkg := range vulnerableRSCPackages {
		if version := pkgJson.getDependencyVersion(pkg); version != "" {
			if fixedVersion := getFixedVersionForRSC(version); fixedVersion != "" {
				vulnerabilities = append(vulnerabilities, SecurityVulnerabilityInfo{
					Package:        pkg,
					CurrentVersion: version,
					FixedVersion:   fixedVersion,
					Location:       location,
					Type:           VulnTypeRSCPackage,
				})
			}
		}
	}

	// Check for Next.js vulnerability
	if version := pkgJson.getDependencyVersion("next"); version != "" {
		if fixedVersion := getFixedVersionForNextJS(version); fixedVersion != "" {
			vulnerabilities = append(vulnerabilities, SecurityVulnerabilityInfo{
				Package:        "next",
				CurrentVersion: version,
				FixedVersion:   fixedVersion,
				Location:       location,
				Type:           VulnTypeNextJS,
			})
		}
	}

	// Check for React Router vulnerability (uses RSC in v7+)
	if version := pkgJson.getDependencyVersion("react-router"); version != "" {
		if fixedVersion := getFixedVersionForReactRouter(version); fixedVersion != "" {
			vulnerabilities = append(vulnerabilities, SecurityVulnerabilityInfo{
				Package:        "react-router",
				CurrentVersion: version,
				FixedVersion:   fixedVersion,
				Location:       location,
				Type:           VulnTypeReactRouter,
			})
		}
	}

	// Also check @react-router/dev which is commonly used
	if version := pkgJson.getDependencyVersion("@react-router/dev"); version != "" {
		if fixedVersion := getFixedVersionForReactRouter(version); fixedVersion != "" {
			vulnerabilities = append(vulnerabilities, SecurityVulnerabilityInfo{
				Package:        "@react-router/dev",
				CurrentVersion: version,
				FixedVersion:   fixedVersion,
				Location:       location,
				Type:           VulnTypeReactRouter,
			})
		}
	}

	// Check for other RSC-affected frameworks
	for _, framework := range rscAffectedFrameworks {
		if version := pkgJson.getDependencyVersion(framework); version != "" {
			vulnerabilities = append(vulnerabilities, SecurityVulnerabilityInfo{
				Package:        framework,
				CurrentVersion: version,
				FixedVersion:   "latest",
				Location:       location,
				Type:           VulnTypeRSCFramework,
			})
		}
	}

	return vulnerabilities
}

// getFixedVersionForRSC returns the fixed version for a given vulnerable React version
// Returns empty string if the version is not vulnerable or already fixed
func getFixedVersionForRSC(versionStr string) string {
	cleanVersion := cleanVersionString(versionStr)
	if cleanVersion == "" {
		return ""
	}

	v, err := semver.NewVersion(cleanVersion)
	if err != nil {
		return ""
	}

	// Only React 19.x versions are affected
	if v.Major() != 19 {
		return ""
	}

	minorKey := fmt.Sprintf("%d.%d", v.Major(), v.Minor())
	fixedVersionStr, exists := rscFixedVersions[minorKey]
	if !exists {
		// Unknown minor version (e.g., 19.3+), assume safe
		return ""
	}

	fixedVersion, err := semver.NewVersion(fixedVersionStr)
	if err != nil {
		return ""
	}

	if v.LessThan(fixedVersion) {
		return fixedVersionStr
	}

	return ""
}

// getFixedVersionForNextJS returns the fixed version for a given vulnerable Next.js version
// Returns empty string if the version is not vulnerable or already fixed
func getFixedVersionForNextJS(versionStr string) string {
	cleanVersion := cleanVersionString(versionStr)
	if cleanVersion == "" {
		return ""
	}

	v, err := semver.NewVersion(cleanVersion)
	if err != nil {
		return ""
	}

	// Next.js 13.3+ through 16.0.x are affected
	major := v.Major()
	if major < 13 || major > 16 {
		return ""
	}

	// Next.js 13.0, 13.1, 13.2 are not affected (RSC wasn't fully available)
	if major == 13 && v.Minor() < 3 {
		return ""
	}

	minorKey := fmt.Sprintf("%d.%d", major, v.Minor())
	fixedVersionStr, exists := nextjsFixedVersions[minorKey]
	if !exists {
		// Unknown minor version (e.g., 16.1+), assume safe
		return ""
	}

	fixedVersion, err := semver.NewVersion(fixedVersionStr)
	if err != nil {
		return ""
	}

	if v.LessThan(fixedVersion) {
		return fixedVersionStr
	}

	return ""
}

// getFixedVersionForReactRouter returns the fixed version for a given vulnerable React Router version
// Returns empty string if the version is not vulnerable or already fixed
func getFixedVersionForReactRouter(versionStr string) string {
	cleanVersion := cleanVersionString(versionStr)
	if cleanVersion == "" {
		return ""
	}

	v, err := semver.NewVersion(cleanVersion)
	if err != nil {
		return ""
	}

	// React Router 7.x with RSC support is affected
	if v.Major() != 7 {
		return ""
	}

	minorKey := fmt.Sprintf("%d.%d", v.Major(), v.Minor())
	fixedVersionStr, exists := reactRouterFixedVersions[minorKey]
	if !exists {
		// Unknown minor version (e.g., 7.7+), assume safe
		return ""
	}

	fixedVersion, err := semver.NewVersion(fixedVersionStr)
	if err != nil {
		return ""
	}

	if v.LessThan(fixedVersion) {
		return fixedVersionStr
	}

	return ""
}

// cleanVersionString removes semver range operators to extract the base version
func cleanVersionString(version string) string {
	version = strings.TrimSpace(version)

	// Handle common version prefixes
	prefixes := []string{"^", "~", ">=", "<=", ">", "<", "=", "v"}
	for _, prefix := range prefixes {
		version = strings.TrimPrefix(version, prefix)
	}

	// Handle ranges like "19.0.0 - 19.0.2" - take the first version
	if parts := strings.Split(version, " - "); len(parts) > 0 {
		version = strings.TrimSpace(parts[0])
	}

	// Handle OR ranges like "19.0.0 || 19.1.0" - take the first version
	if parts := strings.Split(version, "||"); len(parts) > 0 {
		version = strings.TrimSpace(parts[0])
	}

	// Handle AND ranges like ">=19.0.0 <19.1.0" - take the first version
	if parts := strings.Fields(version); len(parts) > 0 {
		version = parts[0]
		// Clean prefixes again after splitting
		for _, prefix := range prefixes {
			version = strings.TrimPrefix(version, prefix)
		}
	}

	return version
}

// warnAndMitigateVulnerabilities logs warnings and returns override commands for vulnerable packages
func (p *NodeProvider) warnAndMitigateVulnerabilities(ctx *generate.GenerateContext, vulnerabilities []SecurityVulnerabilityInfo) map[string]string {
	if len(vulnerabilities) == 0 {
		return nil
	}

	// Categorize vulnerabilities by type for clearer logging
	var rscVulns, nextVulns, routerVulns, frameworkVulns []SecurityVulnerabilityInfo
	for _, v := range vulnerabilities {
		switch v.Type {
		case VulnTypeNextJS:
			nextVulns = append(nextVulns, v)
		case VulnTypeReactRouter:
			routerVulns = append(routerVulns, v)
		case VulnTypeRSCFramework:
			frameworkVulns = append(frameworkVulns, v)
		default:
			rscVulns = append(rscVulns, v)
		}
	}

	overrides := make(map[string]string)

	// Log common CVE info header
	ctx.Logger.LogWarn("⚠️  SECURITY: React Server Components vulnerabilities detected")
	ctx.Logger.LogWarn("   CVE-2025-55182 (Critical RCE), CVE-2025-55183, CVE-2025-55184, CVE-2025-67779")
	ctx.Logger.LogWarn("   See: https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components")

	if len(rscVulns) > 0 {
		ctx.Logger.LogWarn("   Vulnerable RSC packages:")
		for _, vuln := range rscVulns {
			p.logVulnerability(ctx, vuln)
			overrides[vuln.Package] = vuln.FixedVersion
		}
	}

	if len(nextVulns) > 0 {
		ctx.Logger.LogWarn("   Vulnerable Next.js versions:")
		for _, vuln := range nextVulns {
			p.logVulnerability(ctx, vuln)
			overrides[vuln.Package] = vuln.FixedVersion
		}
	}

	if len(routerVulns) > 0 {
		ctx.Logger.LogWarn("   Vulnerable React Router versions:")
		for _, vuln := range routerVulns {
			p.logVulnerability(ctx, vuln)
			overrides[vuln.Package] = vuln.FixedVersion
		}
	}

	if len(frameworkVulns) > 0 {
		ctx.Logger.LogWarn("   RSC frameworks requiring RSC package upgrades:")
		for _, vuln := range frameworkVulns {
			locationInfo := ""
			if vuln.Location != "root" {
				locationInfo = fmt.Sprintf(" (in %s)", vuln.Location)
			}
			ctx.Logger.LogWarn("   - %s@%s detected%s - upgrading RSC dependencies",
				vuln.Package, vuln.CurrentVersion, locationInfo)
		}
		// For framework vulnerabilities, ensure RSC packages are upgraded to latest safe versions
		overrides["react-server-dom-webpack"] = rscFixedVersions["19.2"]
		overrides["react-server-dom-parcel"] = rscFixedVersions["19.2"]
		overrides["react-server-dom-turbopack"] = rscFixedVersions["19.2"]
	}

	ctx.Logger.LogWarn("   Railpack will automatically install patched versions during build")

	return overrides
}

func (p *NodeProvider) logVulnerability(ctx *generate.GenerateContext, vuln SecurityVulnerabilityInfo) {
	locationInfo := ""
	if vuln.Location != "root" {
		locationInfo = fmt.Sprintf(" (in %s)", vuln.Location)
	}
	ctx.Logger.LogWarn("   - %s@%s → %s%s",
		vuln.Package, vuln.CurrentVersion, vuln.FixedVersion, locationInfo)
}

// getOverrideCommands generates package manager specific commands to override vulnerable packages
func (p *NodeProvider) getOverrideCommands(ctx *generate.GenerateContext, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return nil
	}

	// Build override entries for the node script
	var overrideEntries []string
	for pkg, version := range overrides {
		overrideEntries = append(overrideEntries, fmt.Sprintf(`"%s": "%s"`, pkg, version))
	}
	overridesJSON := "{" + strings.Join(overrideEntries, ", ") + "}"

	// Determine the field name based on package manager
	var fieldPath string
	switch p.packageManager {
	case PackageManagerPnpm:
		// pnpm uses pnpm.overrides in package.json
		fieldPath = "pnpm.overrides"
	case PackageManagerYarn1, PackageManagerYarnBerry:
		// Yarn uses resolutions in package.json
		fieldPath = "resolutions"
	default:
		// npm and bun use overrides in package.json
		fieldPath = "overrides"
	}

	// Use node to safely modify package.json with proper JSON handling
	// This handles scoped packages (@scope/pkg) and nested paths correctly
	// Using single quotes for shell to avoid escaping issues with JSON
	nodeScript := fmt.Sprintf(
		`node -e 'const fs=require("fs");const p=JSON.parse(fs.readFileSync("package.json"));const o=%s;const k="%s".split(".");let t=p;for(let i=0;i<k.length-1;i++){if(!t[k[i]])t[k[i]]={};t=t[k[i]];}t[k[k.length-1]]={...t[k[k.length-1]],...o};fs.writeFileSync("package.json",JSON.stringify(p,null,2)+"\n");'`,
		overridesJSON,
		fieldPath,
	)

	return []string{nodeScript}
}

// mitigateRSCVulnerabilities checks for vulnerable packages and adds override commands to the install step
func (p *NodeProvider) mitigateRSCVulnerabilities(ctx *generate.GenerateContext, install *generate.CommandStepBuilder) {
	vulnerabilities := p.checkSecurityVulnerabilities(ctx)
	if len(vulnerabilities) == 0 {
		return
	}

	overrides := p.warnAndMitigateVulnerabilities(ctx, vulnerabilities)
	if len(overrides) == 0 {
		return
	}

	commands := p.getOverrideCommands(ctx, overrides)
	for _, cmd := range commands {
		install.AddCommand(plan.NewExecCommand(cmd))
	}
}
