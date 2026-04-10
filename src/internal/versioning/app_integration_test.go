package versioning

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type fakeGit struct {
	isRepo      bool
	fetchTagsErr error
	tagsErr     error
	tags        []string
	createTagErr error
	stageErr    error
	commitErr   error
	pushErr     error
	pushTagErr  error

	fetchTagsCalls []string
	createTagCalls []string
	stageCalls     [][]string
	commitCalls    []string
	pushCalls      []string
	pushTagCalls   []string
}

func (f *fakeGit) IsRepo(root string) bool {
	return f.isRepo
}

func (f *fakeGit) FetchTags(root string) error {
	f.fetchTagsCalls = append(f.fetchTagsCalls, root)
	return f.fetchTagsErr
}

func (f *fakeGit) Tags(root string) ([]string, error) {
	return f.tags, f.tagsErr
}

func (f *fakeGit) CreateTag(root, tag string) error {
	f.createTagCalls = append(f.createTagCalls, tag)
	return f.createTagErr
}

func (f *fakeGit) StageFiles(root string, files []string) error {
	copied := append([]string(nil), files...)
	f.stageCalls = append(f.stageCalls, copied)
	return f.stageErr
}

func (f *fakeGit) Commit(root, message string) error {
	f.commitCalls = append(f.commitCalls, message)
	return f.commitErr
}

func (f *fakeGit) Push(root string) error {
	f.pushCalls = append(f.pushCalls, root)
	return f.pushErr
}

func (f *fakeGit) PushTag(root, tag string) error {
	f.pushTagCalls = append(f.pushTagCalls, tag)
	return f.pushTagErr
}

func TestResolveReturnsErrorOutsideGitRepoWhenVersionFilesExist(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "meta.yml"), "name: demo\nversion: 1.2.3\n")

	app := NewApp(&fakeGit{}, dir)

	_, err := app.Resolve()
	if err != ErrNotGitRepo {
		t.Fatalf("got %v", err)
	}
}

func TestResolveSyncsLocalFilesToLatestGitVersion(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "package.json"), "{\n  \"name\": \"demo\",\n  \"version\": \"1.0.0\"\n}\n")
	writeTestFile(t, filepath.Join(dir, "meta.yml"), "name: demo\nversion: 1.0.0\ntag_pattern: release/$MAJOR.$MINOR.$PATCH\n")

	git := &fakeGit{
		isRepo: true,
		tags:   []string{"junk", "v1.2.4", "v1.2.3"},
	}
	app := NewApp(git, dir)

	result, err := app.Resolve()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.LocalVersion == nil || result.LocalVersion.String() != "1.0.0" {
		t.Fatalf("local version = %#v", result.LocalVersion)
	}
	if result.GitVersion == nil || result.GitVersion.String() != "1.2.4" {
		t.Fatalf("git version = %#v", result.GitVersion)
	}
	if result.ResolvedVersion == nil || result.ResolvedVersion.String() != "1.2.4" {
		t.Fatalf("resolved version = %#v", result.ResolvedVersion)
	}

	files, err := LoadVersionFiles(dir)
	if err != nil {
		t.Fatalf("reload files: %v", err)
	}
	if files.PrimaryVersion() != "1.2.4" {
		t.Fatalf("synced version = %s", files.PrimaryVersion())
	}
	if len(git.fetchTagsCalls) != 1 {
		t.Fatalf("fetch tags calls = %d", len(git.fetchTagsCalls))
	}
}

func TestApplySyncsFilesAndCreatesTagFromPattern(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "package.json"), "{\n  \"version\": \"1.0.0\"\n}\n")
	writeTestFile(t, filepath.Join(dir, "meta.yml"), "name: demo\nversion: 1.0.0\ntag_pattern: release/$MAJOR.$MINOR.$PATCH\n")

	git := &fakeGit{}
	app := NewApp(git, dir)

	pre := 1
	tag, err := app.Apply(Version{Major: 2, Minor: 3, Patch: 4, Pre: &pre})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tag != "release/2.3.4-1" {
		t.Fatalf("tag = %s", tag)
	}
	if !reflect.DeepEqual(git.createTagCalls, []string{"release/2.3.4-1"}) {
		t.Fatalf("create tag calls = %#v", git.createTagCalls)
	}

	files, err := LoadVersionFiles(dir)
	if err != nil {
		t.Fatalf("reload files: %v", err)
	}
	if files.PrimaryVersion() != "2.3.4-1" {
		t.Fatalf("synced version = %s", files.PrimaryVersion())
	}
}

func TestPushStagesVersionFilesCommitsAndPushesTag(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "package.json"), "{\n  \"version\": \"1.2.3\"\n}\n")
	writeTestFile(t, filepath.Join(dir, "meta.yml"), "name: demo\nversion: 1.2.3\n")

	git := &fakeGit{}
	app := NewApp(git, dir)

	if err := app.Push("v1.2.3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStage := [][]string{{
		filepath.Join(dir, "package.json"),
		filepath.Join(dir, "meta.yml"),
	}}
	if !reflect.DeepEqual(git.stageCalls, expectedStage) {
		t.Fatalf("stage calls = %#v", git.stageCalls)
	}
	if !reflect.DeepEqual(git.commitCalls, []string{"chore: bump version to v1.2.3"}) {
		t.Fatalf("commit calls = %#v", git.commitCalls)
	}
	if len(git.pushCalls) != 1 {
		t.Fatalf("push calls = %d", len(git.pushCalls))
	}
	if !reflect.DeepEqual(git.pushTagCalls, []string{"v1.2.3"}) {
		t.Fatalf("push tag calls = %#v", git.pushTagCalls)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
