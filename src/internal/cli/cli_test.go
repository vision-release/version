package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"version/internal/versioning"
)

type fakeGit struct {
	isRepo        bool
	tags          []string
	createTagCalls []string
	pushTagCalls  []string
	pushCalls     int
}

func (f *fakeGit) IsRepo(root string) bool { return f.isRepo }

func (f *fakeGit) FetchTags(root string) error { return nil }

func (f *fakeGit) Tags(root string) ([]string, error) { return f.tags, nil }

func (f *fakeGit) CreateTag(root, tag string) error {
	f.createTagCalls = append(f.createTagCalls, tag)
	return nil
}

func (f *fakeGit) StageFiles(root string, files []string) error { return nil }

func (f *fakeGit) Commit(root, message string) error { return nil }

func (f *fakeGit) Push(root string) error {
	f.pushCalls++
	return nil
}

func (f *fakeGit) PushTag(root, tag string) error {
	f.pushTagCalls = append(f.pushTagCalls, tag)
	return nil
}

func TestRunInitUsesCurrentDirectoryNameByDefault(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prev)
	})

	app := versioning.NewApp(&fakeGit{}, dir)
	var stdout bytes.Buffer

	if err := runInit(app, strings.NewReader("\n"), &stdout); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(stdout.String(), "project name ["+filepath.Base(dir)+"]: ") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	files, err := versioning.LoadVersionFiles(dir)
	if err != nil {
		t.Fatalf("load version files: %v", err)
	}
	if files.PrimaryVersion() != "0.0.0" {
		t.Fatalf("version = %s", files.PrimaryVersion())
	}
}

func TestRunResolvePrintsResolvedInformation(t *testing.T) {
	dir := t.TempDir()
	writeCLIFile(t, filepath.Join(dir, "meta.yml"), "name: demo\nversion: 1.0.0\n")

	app := versioning.NewApp(&fakeGit{isRepo: true, tags: []string{"v1.0.1"}}, dir)
	var stdout bytes.Buffer

	if err := runResolve(app, &stdout); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	for _, want := range []string{
		"git repo: true",
		"local version: 1.0.0",
		"git version: 1.0.1",
		"resolved version: 1.0.1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout missing %q: %q", want, got)
		}
	}
}

func TestRunBumpStopsWhenCreateTagIsRejected(t *testing.T) {
	dir := t.TempDir()
	app := versioning.NewApp(&fakeGit{isRepo: true, tags: []string{"v1.2.3"}}, dir)
	var stdout bytes.Buffer

	git := app // keep app reference only for invocation symmetry
	_ = git

	err := runBump(app, strings.NewReader("n\n"), &stdout, "patch", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := stdout.String()
	if !strings.Contains(got, "v1.2.4\n") {
		t.Fatalf("stdout = %q", got)
	}
	if !strings.Contains(got, "create tag v1.2.4? [y/N]: ") {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunBumpForceCreatesAndPushesTag(t *testing.T) {
	dir := t.TempDir()
	git := &fakeGit{isRepo: true, tags: []string{"v1.2.3"}}
	app := versioning.NewApp(git, dir)
	var stdout bytes.Buffer

	if err := runBump(app, strings.NewReader(""), &stdout, "patch", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(git.createTagCalls) != 1 || git.createTagCalls[0] != "v1.2.4" {
		t.Fatalf("create tag calls = %#v", git.createTagCalls)
	}
	if len(git.pushTagCalls) != 1 || git.pushTagCalls[0] != "v1.2.4" {
		t.Fatalf("push tag calls = %#v", git.pushTagCalls)
	}
	if git.pushCalls != 0 {
		t.Fatalf("push calls = %d", git.pushCalls)
	}
	if !strings.Contains(stdout.String(), "pushed v1.2.4") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func writeCLIFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
