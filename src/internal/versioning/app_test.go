package versioning

import "testing"

func TestNextVersionPatchFromPreRelease(t *testing.T) {
	pre := 4
	base := Version{Major: 1, Minor: 2, Patch: 3, Pre: &pre}

	got, err := nextVersion(base, true, "patch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.String() != "1.2.4" {
		t.Fatalf("got %s", got.String())
	}
}

func TestNextVersionPrePatchSequence(t *testing.T) {
	got, err := nextVersion(Version{Major: 2, Minor: 3, Patch: 3}, true, "prepatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.String() != "2.3.4-0" {
		t.Fatalf("got %s", got.String())
	}

	pre := 0
	got, err = nextVersion(Version{Major: 2, Minor: 3, Patch: 4, Pre: &pre}, true, "prepatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.String() != "2.3.4-1" {
		t.Fatalf("got %s", got.String())
	}
}

func TestVersionCompareStableBeforePre(t *testing.T) {
	pre := 0
	stable := Version{Major: 1, Minor: 0, Patch: 0}
	prerelease := Version{Major: 1, Minor: 0, Patch: 0, Pre: &pre}

	if stable.Compare(prerelease) != -1 {
		t.Fatalf("expected stable to sort before prerelease")
	}
}
