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
		if len(args) > 1 {
			if printCommandHelp(stdout, args[1]) {
				return nil
			}
			printHelp(stderr)
			return fmt.Errorf("unknown command: %s", args[1])
		}
		printHelp(stdout)
		return nil
	case "init":
		if wantsHelp(args[1:]) {
			printCommandHelp(stdout, "init")
			return nil
		}
		return runInit(app, stdin, stdout)
	case "current":
		if wantsHelp(args[1:]) {
			printCommandHelp(stdout, "current")
			return nil
		}
		return runCurrent(app, stdout)
	case "resolve":
		if wantsHelp(args[1:]) {
			printCommandHelp(stdout, "resolve")
			return nil
		}
		return runResolve(app, stdout)
	case "url":
		if wantsHelp(args[1:]) {
			printCommandHelp(stdout, "url")
			return nil
		}
		return runURL(app, stdout)
	case "patch", "minor", "major", "prepatch", "preminor", "premajor":
		if wantsHelp(args[1:]) {
			printCommandHelp(stdout, args[0])
			return nil
		}
		yes := len(args) > 1 && args[1] == "-y"
		return runBump(app, stdin, stdout, args[0], yes)
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

func runURL(app *versioning.App, stdout io.Writer) error {
	result, err := app.RepositoryURLs()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stdout, "git repository url: %s\n", result.GitURL)
	_, _ = fmt.Fprintf(stdout, "repository link: %s\n", result.Repository)
	_, _ = fmt.Fprintf(stdout, "pipeline link: %s\n", result.PipelineURL)
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

func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			return true
		}
	}
	return false
}

func printCommandHelp(w io.Writer, command string) bool {
	switch strings.ToLower(command) {
	case "help":
		printHelp(w)
	case "init":
		_, _ = fmt.Fprintln(w, "version init")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Create meta.yml in the current directory.")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Behavior:")
		_, _ = fmt.Fprintln(w, "  Prompts for a project name.")
		_, _ = fmt.Fprintln(w, "  Uses the current folder name as the default.")
		_, _ = fmt.Fprintln(w, "  Writes version: 0.0.0 and the default tag pattern.")
		return true
	case "current":
		_, _ = fmt.Fprintln(w, "version current")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Print the current local version from package.json or meta.yml.")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Notes:")
		_, _ = fmt.Fprintln(w, "  Reads local files only.")
		_, _ = fmt.Fprintln(w, "  Works outside a Git repository.")
		return true
	case "resolve":
		_, _ = fmt.Fprintln(w, "version resolve")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Fetch tags, compare local files with Git, and print the resolved state.")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Notes:")
		_, _ = fmt.Fprintln(w, "  Syncs package.json and meta.yml when both exist.")
		_, _ = fmt.Fprintln(w, "  Requires a Git repository.")
		return true
	case "url":
		_, _ = fmt.Fprintln(w, "version url")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Read the origin remote and print repository URLs for GitHub or GitLab.")
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Output:")
		_, _ = fmt.Fprintln(w, "  The configured Git remote URL.")
		_, _ = fmt.Fprintln(w, "  The public repository link.")
		_, _ = fmt.Fprintln(w, "  The pipeline page link.")
		return true
	case "patch":
		printBumpHelp(w, "patch", "Create the next patch version.", "v1.2.3 -> v1.2.4, v1.2.3-4 -> v1.2.4")
		return true
	case "minor":
		printBumpHelp(w, "minor", "Create the next minor version.", "v1.2.3 -> v1.3.0")
		return true
	case "major":
		printBumpHelp(w, "major", "Create the next major version.", "v1.2.3 -> v2.0.0")
		return true
	case "prepatch":
		printBumpHelp(w, "prepatch", "Create or continue a numeric prerelease for the next patch version.", "v1.2.3 -> v1.2.4-0, v1.2.4-0 -> v1.2.4-1")
		return true
	case "preminor":
		printBumpHelp(w, "preminor", "Create or continue a numeric prerelease for the next minor version.", "v1.2.3 -> v1.3.0-0")
		return true
	case "premajor":
		printBumpHelp(w, "premajor", "Create or continue a numeric prerelease for the next major version.", "v1.2.3 -> v2.0.0-0")
		return true
	default:
		return false
	}

	return false
}

func printBumpHelp(w io.Writer, command, description, example string) {
	_, _ = fmt.Fprintf(w, "version %s [--help] [-y]\n", command)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, description)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintf(w, "Example: %s\n", example)
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Options:")
	_, _ = fmt.Fprintln(w, "  -y    Skip prompts, create the tag, and push it immediately.")
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
	_, _ = fmt.Fprintln(w, "  help [command]")
	_, _ = fmt.Fprintln(w, "             Show global help or help for a specific command.")
	_, _ = fmt.Fprintln(w, "  init       Create meta.yml and ask for a project name.")
	_, _ = fmt.Fprintln(w, "             The default project name is the current folder name.")
	_, _ = fmt.Fprintln(w, "  current    Read the current local version from JSON files only.")
	_, _ = fmt.Fprintln(w, "  resolve    Fetch tags, compare local version and latest Git tag,")
	_, _ = fmt.Fprintln(w, "             then print the resolved information. If package.json")
	_, _ = fmt.Fprintln(w, "             and meta.yml both exist, resolve syncs both files.")
	_, _ = fmt.Fprintln(w, "  url        Read the origin remote and print the Git URL, public")
	_, _ = fmt.Fprintln(w, "             repository link, and pipeline link for GitHub or GitLab.")
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
	_, _ = fmt.Fprintln(w, "  Use --help after any command, for example: version prepatch --help")
}
