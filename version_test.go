package version

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewVersion(t *testing.T) {
	cases := []struct {
		version string
		err     bool
	}{
		{"1.2.3", false},
		{"1.0", false},
		{"1", false},
		{"1.2.beta", true},
		{"foo", true},
		{"1.2-5", false},
		{"1.2-beta.5", false},
		{"\n1.2", true},
		{"1.2.0-x.Y.0+metadata", false},
		{"1.2.0-x.Y.0+metadata-width-hypen", false},
		{"1.2.3-rc1-with-hypen", false},
		{"1.2.3.4", true},
	}

	for _, tc := range cases {
		_, err := NewVersion(tc.version)
		if tc.err && err == nil {
			t.Fatalf("expected error for version: %s", tc.version)
		} else if !tc.err && err != nil {
			t.Fatalf("error for version %s: %s", tc.version, err)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2.3", "1.4.5", -1},
		{"1.2-beta", "1.2-beta", 0},
		{"1.2", "1.1.4", 1},
		{"1.2", "1.2-beta", 1},
		{"1.2+foo", "1.2+beta", 0},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.Compare(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s <=> %s\nexpected: %d\nactual: %d",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestComparePreReleases(t *testing.T) {
	cases := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.2-beta.2", "1.2-beta.2", 0},
		{"1.2-beta.1", "1.2-beta.2", -1},
		{"3.2-alpha.1", "3.2-alpha", 1},
		{"1.2-beta.2", "1.2-beta.1", 1},
		{"1.2-beta", "1.2-beta.3", -1},
		{"1.2-alpha", "1.2-beta.3", -1},
		{"1.2-beta", "1.2-alpha.3", 1},
		{"3.0-alpha.3", "3.0-rc.1", -1},
		{"3.0-alpha3", "3.0-rc1", -1},
		{"3.0-alpha.1", "3.0-alpha.beta", -1},
		{"5.4-alpha", "5.4-alpha.beta", 1},
	}

	for _, tc := range cases {
		v1, err := NewVersion(tc.v1)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		v2, err := NewVersion(tc.v2)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v1.Compare(v2)
		expected := tc.expected
		if actual != expected {
			t.Fatalf(
				"%s <=> %s\nexpected: %d\nactual: %d",
				tc.v1, tc.v2,
				expected, actual)
		}
	}
}

func TestVersionMetadata(t *testing.T) {
	cases := []struct {
		version  string
		expected string
	}{
		{"1.2.3", ""},
		{"1.2-beta", ""},
		{"1.2.0-x.Y.0", ""},
		{"1.2.0-x.Y.0+metadata", "metadata"},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc.version)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v.Metadata()
		expected := tc.expected
		if actual != expected {
			t.Fatalf("expected: %s\nactual: %s", expected, actual)
		}
	}
}

func TestVersionPrerelease(t *testing.T) {
	cases := []struct {
		version  string
		expected string
	}{
		{"1.2.3", ""},
		{"1.2-beta", "beta"},
		{"1.2.0-x.Y.0", "x.Y.0"},
		{"1.2.0-x.Y.0+metadata", "x.Y.0"},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc.version)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v.Prerelease()
		expected := tc.expected
		if actual != expected {
			t.Fatalf("expected: %s\nactual: %s", expected, actual)
		}
	}
}

func TestVersionSegments(t *testing.T) {
	cases := []struct {
		version  string
		expected []int
	}{
		{"1.2.3", []int{1, 2, 3}},
		{"1.2-beta", []int{1, 2, 0}},
		{"1-x.Y.0", []int{1, 0, 0}},
		{"1.2.0-x.Y.0+metadata", []int{1, 2, 0}},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc.version)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v.Segments()
		expected := tc.expected
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expected: %#v\nactual: %#v", expected, actual)
		}
	}
}

func TestVersionString(t *testing.T) {
	cases := [][]string{
		{"1.2.3", "1.2.3"},
		{"1.2-beta", "1.2.0-beta"},
		{"1.2.0-x.Y.0", "1.2.0-x.Y.0"},
		{"1.2.0-x.Y.0+metadata", "1.2.0-x.Y.0+metadata"},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc[0])
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		actual := v.String()
		expected := tc[1]
		if actual != expected {
			t.Fatalf("expected: %s\nactual: %s", expected, actual)
		}
	}
}

func TestSetPart(t *testing.T) {
	cases := []struct {
		version string
		part    VersionPart
		val     int
		result  string
		err     bool
	}{
		{"1.1.1", MajorPart, 2, "2.1.1", false},
		{"1.1.1", MinorPart, 0, "1.0.1", false},
		{"1.1.1", PatchPart, 10, "1.1.10", false},
		{"1.1.0-beta1", PreReleasePart, 1, "", true},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc.version)
		if err != nil {
			t.Fatalf("error parsing version %s", tc.version)
		}
		err = v.SetPart(tc.part, tc.val)
		if tc.err && err == nil {
			t.Fatalf("expected error for SetPart of part %d in version %s", tc.part, tc.version)
		} else if !tc.err && err != nil {
			t.Fatalf("expected no error for SetPart of part %d in version %s", tc.part, tc.version)
		}
		if !tc.err {
			if v.String() != tc.result {
				t.Fatalf("SetPart %d to %d in %s, expecting: %s\nfound %s", tc.part, tc.val, tc.version, tc.result, v.String())
			}
		}
	}
}

func TestBumpVersion(t *testing.T) {
	cases := []struct {
		version string
		part    VersionPart
		result  string
		err     bool
	}{
		{"1.1.1", MajorPart, "2.1.1", false},
		{"1.1.1", MinorPart, "1.2.1", false},
		{"1.1.1", PatchPart, "1.1.2", false},
		{"2", MinorPart, "2.1.0", false},
		{"2.2", PatchPart, "2.2.1", false},
		{"1.1.0-beta1", MinorPart, "1.2.0-beta1", false},
		{"1.1.0-beta1", PreReleasePart, "", true},
		{"1.1.0-beta1+foo", MetadataPart, "", true},
	}

	for _, tc := range cases {
		v, err := NewVersion(tc.version)
		if err != nil {
			t.Fatalf("error parsing version %s", tc.version)
		}
		err = v.BumpVersion(tc.part)
		if tc.err && err == nil {
			t.Fatalf("expected error for version: %s", tc.version)
		} else if !tc.err && err != nil {
			t.Fatalf("error for version %s: %s", tc.version, err)
		}
		if !tc.err {
			if v.String() != tc.result {
				t.Fatalf("BumpVersion %d, expecting: %s\nfound %s", tc.part, tc.result, v.String())
			}
		}
	}
}

func TestVersionJSON(t *testing.T) {
	type MyStruct struct {
		Ver *Version
	}
	var (
		ver MyStruct
		err error
	)
	jsBytes := []byte(`{"Ver":"1.2.3"}`)
	// data -> struct
	err = json.Unmarshal(jsBytes, &ver)
	if err != nil {
		t.Fatalf("expected: json.Unmarshal to succeed\nactual: failed with error %v", err)
	}
	// struct -> data
	data, err := json.Marshal(&ver)
	if err != nil {
		t.Fatalf("expected: json.Marshal to succeed\nactual: failed with error %v", err)
	}

	if !bytes.Equal(data, jsBytes) {
		t.Fatalf("expected: %s\nactual: %s", jsBytes, data)
	}
}
