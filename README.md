# version

`version` is an open-source Go CLI for teams that want a simple, Git-based way to define versions and release tags without depending on `yarn version`.

It reads versions from local project files, resolves the latest release state from Git tags, updates version files in sync, creates the next release tag, and can push the result to your remote.

## Why Use It

- Replace `yarn version` with a tool that works from Git tags as the source of truth.
- Keep `package.json` and `meta.yml` aligned automatically.
- Support projects with or without Node.js.
- Standardize patch, minor, major, and numeric prerelease flows.
- Customize the Git tag format with a simple pattern.

## What The Tool Manages

The tool supports these version files:

- `package.json`
- `meta.yml`

Behavior:

- If `package.json` exists, it is the primary local version file.
- If `meta.yml` exists, it is also kept in sync.
- If both files exist, `version` updates both.
- If neither file exists, `current` reports that no version file was found.

Git is the source of truth for release history. The tool fetches tags, detects the highest version tag, and uses that to compute the next version.

## Tag Format

By default, tags are created with this pattern:

```text
v$MAJOR.$MINOR.$PATCH
```

Examples:

- `v1.2.3`
- `v1.2.4-0`
- `v2.0.0-3`

You can override the pattern in `meta.yml`:

```yml
name: my-project
version: 1.2.3
tag_pattern: release/$MAJOR.$MINOR.$PATCH
```

That would produce tags like:

- `release/1.2.3`
- `release/1.2.4-0`

## Requirements

- Go installed locally to build the binary
- Git installed and available in `PATH`
- A Git repository for version-changing commands

`current` can still read local version files outside a Git repository, but commands that create new versions depend on Git.

## Installation

### Unix / macOS

Build and install into `~/.local/bin`:

```sh
./install.sh
```

If `~/.local/bin` is not in your `PATH`, add it:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

You can also choose a custom install location:

```sh
PREFIX="$HOME/tools" ./install.sh
```

### Windows PowerShell

Build the executable:

```powershell
./install.ps1
```

This creates `version.exe` in the project root.

### Build Manually

```sh
make build
```

The binary will be written to:

```text
bin/version
```

## Quick Start

### 1. Initialize `meta.yml`

```sh
version init
```

The command asks for a project name and creates:

```yml
name: my-project
version: 0.0.0
tag_pattern: v$MAJOR.$MINOR.$PATCH
```

If the current directory is already a Git repository with matching tags, `init` tries to start from the latest Git version.

### 2. Check The Current Local Version

```sh
version current
```

Example output:

```text
1.4.2
```

This reads local files only. It does not inspect Git history.

### 3. Resolve The Current State

```sh
version resolve
```

Example output:

```text
git repo: true
local version: 1.4.2
git version: 1.4.2
resolved version: 1.4.2
```

`resolve` fetches tags, compares the local files with Git, and syncs `package.json` and `meta.yml` when needed.

### 4. Create The Next Version

Preview and confirm the next patch release:

```sh
version patch
```

The command prints the next tag, asks for confirmation, updates the version files, and creates the Git tag.

## Commands

### `version help`

Show the built-in help screen.

### `version init`

Create `meta.yml` in the current directory.

### `version current`

Print the current local version from `package.json` or `meta.yml`.

### `version resolve`

Fetch tags, detect the latest Git version, compare it with local version files, and print the resolved state.

### `version url`

Read the `origin` remote and print:

- the configured Git repository URL
- the public repository link
- the pipeline page link

Supported providers:

- GitHub: pipeline link points to the repository Actions page
- GitLab: pipeline link points to the repository Pipelines page

### `version patch`

Create the next patch version.

Examples:

- `v1.2.3` -> `v1.2.4`
- `v1.2.3-4` -> `v1.2.4`

### `version minor`

Create the next minor version.

Examples:

- `v1.2.3` -> `v1.3.0`
- `v1.2.3-4` -> `v1.3.0`

### `version major`

Create the next major version.

Examples:

- `v1.2.3` -> `v2.0.0`
- `v1.2.3-4` -> `v2.0.0`

### `version prepatch`

Create or continue a numeric prerelease for the next patch version.

Examples:

- `v2.3.3` -> `v2.3.4-0`
- `v2.3.4-0` -> `v2.3.4-1`

### `version preminor`

Create or continue a numeric prerelease for the next minor version.

Examples:

- `v2.3.3` -> `v2.4.0-0`
- `v2.4.0-0` -> `v2.4.0-1`

### `version premajor`

Create or continue a numeric prerelease for the next major version.

Examples:

- `v2.3.3` -> `v3.0.0-0`
- `v3.0.0-0` -> `v3.0.0-1`

## Interactive And Non-Interactive Mode

By default, version-changing commands are interactive:

```sh
version patch
```

Flow:

1. Print the next tag.
2. Ask whether the tag should be created.
3. Create the tag and update local files.
4. Ask whether the tag should be pushed.

You can skip the prompts with `-y` or `-f`:

```sh
version patch -y
```

With `-y` or `-f`, the tool:

- updates version files
- creates the Git tag
- stages changed version files
- commits them
- pushes the branch
- pushes the tag

## Example Workflows

### Node.js Project With `package.json`

```sh
version current
version resolve
version patch
```

Result:

- `package.json` is updated
- a new Git tag is created
- you can optionally push it immediately

### Generic Project With `meta.yml`

```sh
version init
version current
version preminor
```

This works well for projects that do not use `package.json` but still need a consistent version and tag workflow.

### Keep Both Files In Sync

If your repository contains both `package.json` and `meta.yml`, the tool updates both files to the same version so your package metadata and release tags stay aligned.

## Example `meta.yml`

```yml
name: acme-service
version: 1.7.0
tag_pattern: v$MAJOR.$MINOR.$PATCH
```

Fields:

- `name`: project name
- `version`: current local version
- `tag_pattern`: optional tag template

## Notes And Limitations

- The tool requires Git for `resolve`, `patch`, `minor`, `major`, `prepatch`, `preminor`, and `premajor`.
- Tags are fetched before computing the next version.
- Version files are updated before the tag is pushed.
- The tool uses numeric prerelease suffixes such as `-0`, `-1`, `-2`.
- Local versions are stored without the `v` prefix, while Git tags may include it depending on `tag_pattern`.

## Open Source

This project is published as open source by vision release GmbH.

Creator: Darius Sobczak

Add your preferred `LICENSE` file, such as MIT or Apache-2.0, so the repository has an explicit license.
