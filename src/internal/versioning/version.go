package versioning

import (
	"fmt"
	"strconv"
	"strings"
)

const DefaultTagPattern = "v$MAJOR.$MINOR.$PATCH"

type Version struct {
	Major int
	Minor int
	Patch int
	Pre   *int
}

func ParseVersion(input string) (Version, error) {
	parts := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(input), "v"), "-", 2)
	base := strings.Split(parts[0], ".")
	if len(base) != 3 {
		return Version{}, fmt.Errorf("invalid version: %s", input)
	}

	major, err := strconv.Atoi(base[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(base[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(base[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %w", err)
	}

	version := Version{Major: major, Minor: minor, Patch: patch}
	if len(parts) == 2 {
		pre, err := strconv.Atoi(parts[1])
		if err != nil {
			return Version{}, fmt.Errorf("invalid prerelease version: %w", err)
		}
		version.Pre = &pre
	}

	return version, nil
}

func (v Version) String() string {
	if v.Pre == nil {
		return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	}

	return fmt.Sprintf("%d.%d.%d-%d", v.Major, v.Minor, v.Patch, *v.Pre)
}

func (v Version) FormatTag(pattern string) string {
	result := strings.NewReplacer(
		"$MAJOR", strconv.Itoa(v.Major),
		"$MINOR", strconv.Itoa(v.Minor),
		"$PATCH", strconv.Itoa(v.Patch),
	).Replace(pattern)
	if v.Pre != nil {
		result += fmt.Sprintf("-%d", *v.Pre)
	}
	return result
}

func (v Version) IsPreRelease() bool {
	return v.Pre != nil
}

func (v Version) Compare(other Version) int {
	switch {
	case v.Major != other.Major:
		return compareInt(v.Major, other.Major)
	case v.Minor != other.Minor:
		return compareInt(v.Minor, other.Minor)
	case v.Patch != other.Patch:
		return compareInt(v.Patch, other.Patch)
	case v.Pre == nil && other.Pre == nil:
		return 0
	case v.Pre == nil:
		return -1
	case other.Pre == nil:
		return 1
	default:
		return compareInt(*v.Pre, *other.Pre)
	}
}

func (v Version) Stable() Version {
	return Version{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}
}

func (v Version) NextPatch() Version {
	return Version{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}
}

func (v Version) NextMinor() Version {
	return Version{Major: v.Major, Minor: v.Minor + 1, Patch: 0}
}

func (v Version) NextMajor() Version {
	return Version{Major: v.Major + 1, Minor: 0, Patch: 0}
}

func (v Version) WithPre(n int) Version {
	stable := v.Stable()
	stable.Pre = &n
	return stable
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
