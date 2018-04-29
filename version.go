package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// The compiled regular expression used to test the validity of a version.
var versionRegexp *regexp.Regexp

// The raw regular expression string used for testing the validity
// of a version.
const VersionRegexpRaw string = `([0-9]+(\.[0-9]+){0,2})` +
	`(-([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`(\+([0-9A-Za-z\-]+(\.[0-9A-Za-z\-]+)*))?` +
	`?`

type VersionPart int

const (
	MajorPart VersionPart = iota
	MinorPart
	PatchPart
	PreReleasePart
	MetadataPart
)

var partNames = [...]string{
	"major", "minor", "patch", "prerelease", "metadata",
}

// Version represents a single version.
type Version struct {
	metadata string
	pre      string
	segments []int
	si       int
}

func init() {
	versionRegexp = regexp.MustCompile("^" + VersionRegexpRaw + "$")
}

// NewVersion parses the given version and returns a new
// Version.
func NewVersion(v string) (*Version, error) {
	matches := versionRegexp.FindStringSubmatch(v)
	if matches == nil {
		return nil, fmt.Errorf("Malformed version: %s", v)
	}

	segmentsStr := strings.Split(matches[1], ".")
	segments := make([]int, len(segmentsStr), 3)
	si := 0
	for i, str := range segmentsStr {
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, fmt.Errorf(
				"Error parsing version: %s", err)
		}

		segments[i] = int(val)
		si += 1
	}
	for i := len(segments); i < 3; i++ {
		segments = append(segments, 0)
	}

	return &Version{
		metadata: matches[7],
		pre:      matches[4],
		segments: segments,
		si:       si,
	}, nil
}

// Must is a helper that wraps a call to a function returning (*Version, error)
// and panics if error is non-nil.
func Must(v *Version, err error) *Version {
	if err != nil {
		panic(err)
	}

	return v
}

// Compare compares this version to another version. This
// returns -1, 0, or 1 if this version is smaller, equal,
// or larger than the other version, respectively.
//
// If you want boolean results, use the LessThan, Equal,
// or GreaterThan methods.
func (v *Version) Compare(other *Version) int {
	// A quick, efficient equality check
	if v.String() == other.String() {
		return 0
	}

	segmentsSelf := v.Segments()
	segmentsOther := other.Segments()

	// If the segments are the same, we must compare on prerelease info
	if reflect.DeepEqual(segmentsSelf, segmentsOther) {
		preSelf := v.Prerelease()
		preOther := other.Prerelease()
		if preSelf == "" && preOther == "" {
			return 0
		}
		if preSelf == "" {
			return 1
		}
		if preOther == "" {
			return -1
		}

		return comparePrereleases(preSelf, preOther)
	}

	// Compare the segments
	for i := 0; i < len(segmentsSelf); i++ {
		lhs := segmentsSelf[i]
		rhs := segmentsOther[i]

		if lhs == rhs {
			continue
		} else if lhs < rhs {
			return -1
		} else {
			return 1
		}
	}

	panic("should not be reached")
}

func comparePart(preSelf string, preOther string) int {
	if preSelf == preOther {
		return 0
	}

	// if a part is empty, we use the other to decide
	if preSelf == "" {
		_, notIsNumeric := strconv.ParseInt(preOther, 10, 64)
		if notIsNumeric == nil {
			return -1
		}
		return 1
	}

	if preOther == "" {
		_, notIsNumeric := strconv.ParseInt(preSelf, 10, 64)
		if notIsNumeric == nil {
			return 1
		}
		return -1
	}

	if preSelf > preOther {
		return 1
	}

	return -1
}

func comparePrereleases(v string, other string) int {
	// the same pre release!
	if v == other {
		return 0
	}

	// split both pre releases for analyse their parts
	selfPreReleaseMeta := strings.Split(v, ".")
	otherPreReleaseMeta := strings.Split(other, ".")

	selfPreReleaseLen := len(selfPreReleaseMeta)
	otherPreReleaseLen := len(otherPreReleaseMeta)

	biggestLen := otherPreReleaseLen
	if selfPreReleaseLen > otherPreReleaseLen {
		biggestLen = selfPreReleaseLen
	}

	// loop for parts to find the first difference
	for i := 0; i < biggestLen; i = i + 1 {
		partSelfPre := ""
		if i < selfPreReleaseLen {
			partSelfPre = selfPreReleaseMeta[i]
		}

		partOtherPre := ""
		if i < otherPreReleaseLen {
			partOtherPre = otherPreReleaseMeta[i]
		}

		compare := comparePart(partSelfPre, partOtherPre)
		// if parts are equals, continue the loop
		if compare != 0 {
			return compare
		}
	}

	return 0
}

// Equal tests if two versions are equal.
func (v *Version) Equal(o *Version) bool {
	return v.Compare(o) == 0
}

// GreaterThan tests if this version is greater than another version.
func (v *Version) GreaterThan(o *Version) bool {
	return v.Compare(o) > 0
}

// LessThan tests if this version is less than another version.
func (v *Version) LessThan(o *Version) bool {
	return v.Compare(o) < 0
}

// Metadata returns any metadata that was part of the version
// string.
//
// Metadata is anything that comes after the "+" in the version.
// For example, with "1.2.3+beta", the metadata is "beta".
func (v *Version) Metadata() string {
	return v.metadata
}

// Prerelease returns any prerelease data that is part of the version,
// or blank if there is no prerelease data.
//
// Prerelease information is anything that comes after the "-" in the
// version (but before any metadata). For example, with "1.2.3-beta",
// the prerelease information is "beta".
func (v *Version) Prerelease() string {
	return v.pre
}

// Segments returns the numeric segments of the version as a slice.
//
// This excludes any metadata or pre-release information. For example,
// for a version "1.2.3-beta", segments will return a slice of
// 1, 2, 3.
func (v *Version) Segments() []int {
	return v.segments
}

// String returns the full version string included pre-release
// and metadata information.
func (v *Version) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "%d.%d.%d", v.segments[0], v.segments[1], v.segments[2])
	if v.pre != "" {
		fmt.Fprintf(&buf, "-%s", v.pre)
	}
	if v.metadata != "" {
		fmt.Fprintf(&buf, "+%s", v.metadata)
	}

	return buf.String()
}

// SetPart sets MajorPart, MinorPart or PatchPart to v.
func (v *Version) SetPart(part VersionPart, val int) error {
	switch part {
	case MajorPart, MinorPart, PatchPart:
		v.segments[part] = val
	default:
		return fmt.Errorf("unable to set version part %s", partNames[part])
	}
	return nil
}

// BumpVersion does the same as BumpPart but resets all lesser parts to 0.
func (v *Version) BumpVersion(part VersionPart) error {
	const reset = 0

	if part <= PatchPart && (v.pre != "" || v.metadata != "") {
		v.pre = ""
		v.metadata = ""
	}
	if part <= MinorPart {
		if err := v.SetPart(PatchPart, reset); err != nil {
			return fmt.Errorf("unable to reset patch to 0 when bumping minor part in version %q: %s\n", v, err)
		}
	}
	if part <= MajorPart {
		if err := v.SetPart(MinorPart, reset); err != nil {
			return fmt.Errorf("unable to reset minor to 0 when bumping major in part version %q: %s\n", v, err)
		}
	}
	return v.BumpPart(part)
}

// BumpPart increments the indicated part by 1.
// part may be one of: MajorPart, MinorPart or PatchPart
func (v *Version) BumpPart(part VersionPart) (err error) {
	switch part {
	case MajorPart, MinorPart, PatchPart:
		v.segments[part]++
	default:
		err = fmt.Errorf("unable to bump version part %s", partNames[part])
	}
	return
}

// MarshalJSON implements the json packages Marshaler interface
func (v *Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, v.String())), nil
}

// UnmarshalJSON implements the json packages Unmarshaler interface
func (v *Version) UnmarshalJSON(data []byte) (err error) {
	var verStr string
	var nv *Version

	err = json.Unmarshal(data, &verStr)
	if err != nil {
		return
	}

	nv, err = NewVersion(verStr)
	if err != nil {
		return
	}
	*v = *nv

	return
}

// MarshalYAML implements the YAML packages Marshaler interface (gopkg.in/yaml.v2)
func (v *Version) MarshalYAML() (str interface{}, err error) {
	str = v.String()
	return
}

// UnmarshalYAML implements the YAML packages Unmarshaler interface (gopkg.in/yaml.v2)
func (v *Version) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	var (
		verStr string
		nv     *Version
	)
	if err = unmarshal(&verStr); err != nil {
		return
	}
	if nv, err = NewVersion(verStr); err != nil {
		return
	}
	*v = *nv
	return
}
