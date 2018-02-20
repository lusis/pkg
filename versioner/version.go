package version

import (
	"fmt"

	gover "github.com/hashicorp/go-version"
)

const defaultMinVersion = "1.0"
const defaultMaxVersion = "2.0"

// Version is a self-contained go-version Version
type Version struct {
	*gover.Version
}

// Versioner is an interface for a types that supports versioning information
type Versioner interface {
	MinVersion() Version
	MaxVersion() Version
	Deprecated() bool
}

// GetMinVersionFor gets the minimum api version required for a thing
func GetMinVersionFor(a Versioner) Version { return a.MinVersion() }

// GetMaxVersionFor gets the maximum api version required for a thing
func GetMaxVersionFor(a Versioner) Version { return a.MaxVersion() }

// IsDeprecated indicates if a thing is deprecated or not
func IsDeprecated(a Versioner) bool { return a.Deprecated() }

// GenericVersioner is for version checking
// Some operations don't have a response (think DELETE or PUT)
// but we still want to do a version check on ALL functions anyway
// This response simply responds to that
type GenericVersioner struct {
	min        string
	max        string
	deprecated bool
}

// MinVersion returns the minimum version required
func (g *GenericVersioner) MinVersion() Version {
	if g.min == "" {
		g.min = defaultMinVersion
	}
	return MustParse(g.min)
}

// MaxVersion returns the max version required
func (g *GenericVersioner) MaxVersion() Version {
	if g.max == "" {
		g.max = defaultMaxVersion
	}
	return MustParse(g.max)
}

// Deprecated returns if a thing is deprecated
func (g *GenericVersioner) Deprecated() bool { return g.deprecated }

// NewGenericVersioner returns a versioner with the specified constraints
func NewGenericVersioner(minimum, maximum string, isDeprecated bool) *GenericVersioner {
	thing := &GenericVersioner{
		min:        minimum,
		max:        maximum,
		deprecated: isDeprecated,
	}
	return thing
}

// MustParse is a panicing version of NewVersion
func MustParse(v string) Version {
	ver, err := gover.NewVersion(v)
	if err != nil {
		panic("cannot parse version")
	}
	return Version{ver}
}

// CheckSupportedVersion checks a versioner against a provided version string
func CheckSupportedVersion(v Versioner, ver string) error {
	min := GetMinVersionFor(v)
	max := GetMaxVersionFor(v)

	myver, err := gover.NewVersion(ver)
	if err != nil {
		return err
	}
	if myver.Equal(max.Version) || myver.Equal(min.Version) {
		return nil
	}
	if myver.GreaterThan(min.Version) && myver.LessThan(max.Version) {
		return nil
	}
	return fmt.Errorf("Requested version (%s) does not meet the requirements for this type (min: %s, max: %s)",
		myver.String(), min.String(), max.String())
}
