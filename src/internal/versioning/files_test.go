package versioning

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadVersionFilesPrefersPackageJSON(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{\"version\":\"1.2.3\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "meta.yml"), []byte("name: x\nversion: 9.9.9\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := LoadVersionFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if files.PrimaryVersion() != "1.2.3" {
		t.Fatalf("got %s", files.PrimaryVersion())
	}
}

func TestLoadVersionFilesReadsMetaYML(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "meta.yml"), []byte("name: x\nversion: 2.4.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := LoadVersionFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if files.PrimaryVersion() != "2.4.6" {
		t.Fatalf("got %s", files.PrimaryVersion())
	}
}

func TestInitMetaYML(t *testing.T) {
	dir := t.TempDir()

	file, err := InitMetaYML(dir, "demo-project", Version{})
	if err != nil {
		t.Fatal(err)
	}

	if file.Name != "demo-project" {
		t.Fatalf("got %s", file.Name)
	}

	if file.Version.String() != "0.0.0" {
		t.Fatalf("got %s", file.Version.String())
	}

	files, err := LoadVersionFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if files.PrimaryVersion() != "0.0.0" {
		t.Fatalf("got %s", files.PrimaryVersion())
	}
}
