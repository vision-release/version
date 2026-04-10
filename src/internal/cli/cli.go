package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"version/internal/versioning"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	app := versioning.NewApp(versioning.RealGit{}, ".")

	if len(args) == 0 {
		printHelp(stdout)
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "help", "--help", "-h":
		printHelp(stdout)
		return nil
	case "init":
		return runInit(app, stdin, stdout)
	case "current":
		return runCurrent(app, stdout)
	case "resolve":
		return runResolve(app, stdout)
	case "patch", "minor", "major", "prepatch", "preminor", "premajor":
		force := len(args) > 1 && args[1] == "-f"
		return runBump(app, stdin, stdout, args[0], force)
	default:
		printHelp(stderr)
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func runInit(app *versioning.App, stdin io.Reader, stdout io.Writer) error {
	defaultName, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	preset := filepath.Base(defaultName)
	_, _ = fmt.Fprintf(stdout, "project name [%s]: ", preset)

	reader := bufio.NewReader(stdin)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	name := strings.TrimSpace(input)
	if name == "" {
		name = preset
	}

	file, err := app.InitMetaYML(name)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "created %s\n", file.Path)
	return nil
}

func runCurrent(app *versioning.App, stdout io.Writer) error {
	files, err := app.LoadVersionFiles()
	if err != nil {
		if errors.Is(err, versioning.ErrNoVersionFile) {
			_, _ = fmt.Fprintln(stdout, "no version file found")
			return nil
		}

		return err
	}

	_, _ = fmt.Fprintln(stdout, files.PrimaryVersion())
	return nil
}

func runResolve(app *versioning.App, stdout io.Writer) error {
	result, err := app.Resolve()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "git repo: %t\n", result.IsGitRepo)
	if result.LocalVersion == nil {
		_, _ = fmt.Fprintln(stdout, "local version: none")
	} else {
		_, _ = fmt.Fprintf(stdout, "local version: %s\n", result.LocalVersion.String())
	}

	if result.GitVersion == nil {
		_, _ = fmt.Fprintln(stdout, "git version: none")
	} else {
		_, _ = fmt.Fprintf(stdout, "git version: %s\n", result.GitVersion.String())
	}

	if result.ResolvedVersion == nil {
		_, _ = fmt.Fprintln(stdout, "resolved version: none")
	} else {
		_, _ = fmt.Fprintf(stdout, "resolved version: %s\n", result.ResolvedVersion.String())
	}

	return nil
}

func runBump(app *versioning.App, stdin io.Reader, stdout io.Writer, command string, force bool) error {
	next, err := app.Preview(command)
	if err != nil {
		return err
	}

	nextTag := next.FormatTag(app.TagPattern())
	_, _ = fmt.Fprintln(stdout, nextTag)

	reader := bufio.NewReader(stdin)

	if !force {
		_, _ = fmt.Fprintf(stdout, "create tag %s? [y/N]: ", nextTag)
		input, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if strings.ToLower(strings.TrimSpace(input)) != "y" {
			return nil
		}
	}

	tag, err := app.Apply(next)
	if err != nil {
		return err
	}

	if force {
		if err := app.Push(tag); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "pushed %s\n", tag)
		return nil
	}

	_, _ = fmt.Fprintf(stdout, "push tag %s? [y/N]: ", tag)
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if strings.ToLower(strings.TrimSpace(input)) == "y" {
		if err := app.Push(tag); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(stdout, "pushed %s\n", tag)
	}

	return nil
}

func printHelp(w io.Writer) {
	_, _ = fmt.Fprintln(w, "version <command>")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "A Git-based versioning tool for package.json and meta.yml.")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Version files:")
	_, _ = fmt.Fprintln(w, "  package.json is preferred when it exists.")
	_, _ = fmt.Fprintln(w, "  meta.yml is used when package.json does not exist.")
	_, _ = fmt.Fprintln(w, "  If both files exist, their version values are kept in sync.")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Tag format:")
	_, _ = fmt.Fprintln(w, "  Stable tags: v1.2.3 (default pattern: v$MAJOR.$MINOR.$PATCH)")
	_, _ = fmt.Fprintln(w, "  Numeric prerelease tags: v1.2.3-0")
	_, _ = fmt.Fprintln(w, "  Set tag_pattern in meta.yml to customise (e.g. release/$MAJOR.$MINOR.$PATCH)")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Commands:")
	_, _ = fmt.Fprintln(w, "  help       Show this help screen.")
	_, _ = fmt.Fprintln(w, "  init       Create meta.yml and ask for a project name.")
	_, _ = fmt.Fprintln(w, "             The default project name is the current folder name.")
	_, _ = fmt.Fprintln(w, "  current    Read the current local version from JSON files only.")
	_, _ = fmt.Fprintln(w, "  resolve    Fetch tags, compare local version and latest Git tag,")
	_, _ = fmt.Fprintln(w, "             then print the resolved information. If package.json")
	_, _ = fmt.Fprintln(w, "             and meta.yml both exist, resolve syncs both files.")
	_, _ = fmt.Fprintln(w, "  patch      Use the highest Git version and create the next patch tag.")
	_, _ = fmt.Fprintln(w, "             Example: v1.2.3 -> v1.2.4, v1.2.3-4 -> v1.2.4")
	_, _ = fmt.Fprintln(w, "  minor      Use the highest Git version and create the next minor tag.")
	_, _ = fmt.Fprintln(w, "             Example: v1.2.3 -> v1.3.0")
	_, _ = fmt.Fprintln(w, "  major      Use the highest Git version and create the next major tag.")
	_, _ = fmt.Fprintln(w, "             Example: v1.2.3 -> v2.0.0")
	_, _ = fmt.Fprintln(w, "  prepatch   Bump to the next patch version, then start or continue")
	_, _ = fmt.Fprintln(w, "             a numeric prerelease. Example: v1.2.3 -> v1.2.4-0,")
	_, _ = fmt.Fprintln(w, "             v1.2.4-0 -> v1.2.4-1")
	_, _ = fmt.Fprintln(w, "  preminor   Bump to the next minor version, then start or continue")
	_, _ = fmt.Fprintln(w, "             a numeric prerelease. Example: v1.2.3 -> v1.3.0-0")
	_, _ = fmt.Fprintln(w, "  premajor   Bump to the next major version, then start or continue")
	_, _ = fmt.Fprintln(w, "             a numeric prerelease. Example: v1.2.3 -> v2.0.0-0")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Notes:")
	_, _ = fmt.Fprintln(w, "  Git is the source of truth for version history.")
	_, _ = fmt.Fprintln(w, "  Version-changing commands require a Git repository.")
	_, _ = fmt.Fprintln(w, "  current can still read JSON files outside a Git repository.")
}
