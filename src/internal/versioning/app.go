package versioning

import (
	"errors"
	"fmt"
	"slices"
)

type App struct {
	git  GitClient
	root string
}

type ResolveResult struct {
	IsGitRepo       bool
	LocalVersion    *Version
	GitVersion      *Version
	ResolvedVersion *Version
}

func NewApp(git GitClient, root string) *App {
	return &App{git: git, root: root}
}

func (a *App) LoadVersionFiles() (VersionFiles, error) {
	return LoadVersionFiles(a.root)
}

func (a *App) InitMetaYML(name string) (YAMLFile, error) {
	version := Version{}
	if a.git.IsRepo(a.root) {
		if err := a.git.FetchTags(a.root); err == nil {
			if gitVersion, err := a.latestGitVersion(); err == nil && gitVersion != nil {
				version = *gitVersion
			}
		}
	}
	return InitMetaYML(a.root, name, version)
}

func (a *App) Push(tag string) error {
	files, loadErr := a.LoadVersionFiles()
	if loadErr != nil && !errors.Is(loadErr, ErrNoVersionFile) {
		return loadErr
	}

	if loadErr == nil {
		var paths []string
		if files.PackageJSON != nil {
			paths = append(paths, files.PackageJSON.Path)
		}
		if files.MetaYML != nil {
			paths = append(paths, files.MetaYML.Path)
		}
		if len(paths) > 0 {
			if err := a.git.StageFiles(a.root, paths); err != nil {
				return err
			}
			if err := a.git.Commit(a.root, "chore: bump version to "+tag); err != nil {
				return err
			}
			if err := a.git.Push(a.root); err != nil {
				return err
			}
		}
	}

	return a.git.PushTag(a.root, tag)
}

func (a *App) Resolve() (ResolveResult, error) {
	files, err := a.LoadVersionFiles()
	if err != nil && !errors.Is(err, ErrNoVersionFile) {
		return ResolveResult{}, err
	}

	result := ResolveResult{
		IsGitRepo: a.git.IsRepo(a.root),
	}

	if files.PrimaryParsedVersion() != nil {
		result.LocalVersion = files.PrimaryParsedVersion()
	}

	hasFiles := !errors.Is(err, ErrNoVersionFile)
	if !result.IsGitRepo {
		if hasFiles {
			return ResolveResult{}, ErrNotGitRepo
		}
		result.ResolvedVersion = result.LocalVersion
		return result, nil
	}

	if err := a.git.FetchTags(a.root); err != nil {
		return ResolveResult{}, err
	}

	gitVersion, err := a.latestGitVersion()
	if err != nil {
		return ResolveResult{}, err
	}

	result.GitVersion = gitVersion
	result.ResolvedVersion = gitVersion
	if result.ResolvedVersion == nil {
		result.ResolvedVersion = result.LocalVersion
	}

	if result.ResolvedVersion != nil && !errors.Is(err, ErrNoVersionFile) && (files.PackageJSON != nil || files.MetaYML != nil) {
		if syncErr := files.Sync(*result.ResolvedVersion); syncErr != nil {
			return ResolveResult{}, syncErr
		}
	}

	return result, nil
}

func (a *App) Preview(command string) (Version, error) {
	if !a.git.IsRepo(a.root) {
		return Version{}, ErrNotGitRepo
	}

	if err := a.git.FetchTags(a.root); err != nil {
		return Version{}, err
	}

	current, err := a.latestGitVersion()
	if err != nil {
		return Version{}, err
	}

	base := Version{}
	if current != nil {
		base = *current
	}

	return nextVersion(base, current != nil, command)
}

func (a *App) TagPattern() string {
	files, err := a.LoadVersionFiles()
	if err != nil {
		return DefaultTagPattern
	}
	return files.TagPattern()
}

func (a *App) Apply(next Version) (string, error) {
	files, loadErr := a.LoadVersionFiles()
	if loadErr != nil && !errors.Is(loadErr, ErrNoVersionFile) {
		return "", loadErr
	}

	pattern := DefaultTagPattern
	if loadErr == nil {
		if err := files.Sync(next); err != nil {
			return "", err
		}
		pattern = files.TagPattern()
	}

	tag := next.FormatTag(pattern)
	return tag, a.git.CreateTag(a.root, tag)
}

func (a *App) latestGitVersion() (*Version, error) {
	tags, err := a.git.Tags(a.root)
	if err != nil {
		return nil, err
	}

	versions := make([]Version, 0, len(tags))
	for _, tag := range tags {
		version, parseErr := ParseVersion(tag)
		if parseErr != nil {
			continue
		}

		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return nil, nil
	}

	slices.SortFunc(versions, func(a, b Version) int {
		return a.Compare(b)
	})

	version := versions[len(versions)-1]
	return &version, nil
}

func nextVersion(base Version, hasBase bool, command string) (Version, error) {
	if !hasBase {
		switch command {
		case "patch":
			return Version{Major: 0, Minor: 0, Patch: 1}, nil
		case "minor":
			return Version{Major: 0, Minor: 1, Patch: 0}, nil
		case "major":
			return Version{Major: 1, Minor: 0, Patch: 0}, nil
		case "prepatch":
			return Version{Major: 0, Minor: 0, Patch: 1}.WithPre(0), nil
		case "preminor":
			return Version{Major: 0, Minor: 1, Patch: 0}.WithPre(0), nil
		case "premajor":
			return Version{Major: 1, Minor: 0, Patch: 0}.WithPre(0), nil
		default:
			return Version{}, fmt.Errorf("unsupported command: %s", command)
		}
	}

	switch command {
	case "patch":
		if base.IsPreRelease() {
			return base.Stable().NextPatch(), nil
		}
		return base.NextPatch(), nil
	case "minor":
		return base.Stable().NextMinor(), nil
	case "major":
		return base.Stable().NextMajor(), nil
	case "prepatch":
		return nextPreVersion(base, "patch"), nil
	case "preminor":
		return nextPreVersion(base, "minor"), nil
	case "premajor":
		return nextPreVersion(base, "major"), nil
	default:
		return Version{}, fmt.Errorf("unsupported command: %s", command)
	}
}

func nextPreVersion(base Version, mode string) Version {
	if base.IsPreRelease() {
		if impliedPreLevel(base.Stable()) == mode {
			return base.Stable().WithPre(*base.Pre + 1)
		}
		return bumpFromStable(base.Stable(), mode).WithPre(0)
	}

	target := bumpFromStable(base.Stable(), mode)
	return target.WithPre(0)
}

// impliedPreLevel returns the bump level a prerelease version was created for,
// inferred from the shape of its stable base: non-zero patch → patch,
// zero patch + non-zero minor → minor, both zero → major.
func impliedPreLevel(stable Version) string {
	if stable.Patch != 0 {
		return "patch"
	}
	if stable.Minor != 0 {
		return "minor"
	}
	return "major"
}

func bumpFromStable(base Version, mode string) Version {
	switch mode {
	case "patch":
		return base.NextPatch()
	case "minor":
		return base.NextMinor()
	case "major":
		return base.NextMajor()
	default:
		return base
	}
}
