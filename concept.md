version is a CLI tool similar to `yarn version`, implemented in Go.

## Goal

The tool reads Git tags, validates the next possible version, updates local version files, and creates a new tag.

Supported version files:

- `package.json`
- `meta.yml`

## Source Of Truth

Git is the source of truth for version history.

Local JSON files are used as the current local definition:

1. if `package.json` exists, use it first
2. otherwise if `meta.yml` exists, use it
3. otherwise show a message that no version file exists

If both `package.json` and `meta.yml` exist, their version values must stay in sync.

## Tag Format

Tags use this exact format:

- `1.2.3`
- `1.2.3-0`
- `1.2.3-1`

No `v` prefix.
No prerelease identifier like `rc` or `beta`.
Only numeric prerelease suffixes are allowed.

## Commands

- `help`
- `init`
- `current`
- `resolve`
- `patch`
- `minor`
- `major`
- `prepatch`
- `preminor`
- `premajor`

## Command Behavior

### `current`

Reads the current local version from JSON files only.

Order:

1. use `package.json` if it exists
2. otherwise use `meta.yml` if it exists
3. otherwise print a message that no version file exists

`current` does not resolve Git history.

### `init`

Creates `meta.yml`.

It asks the user for the project name and uses the current folder name as the preset.

Initial file:

```yml
name: project-name
version: 0.0.0
```

### `resolve`

`resolve` is an information screen.

It will:

- fetch Git tags
- read the current local version definition
- read the latest Git tag
- revalidate the current state
- show the resolved information

If both `package.json` and `meta.yml` exist, `resolve` updates both files so they stay in sync with the resolved version.

### `patch`

Uses the highest Git version as the source.

If the highest Git version is stable, bump patch:

- `1.2.3` -> `1.2.4`

If the highest Git version is prerelease, remove the prerelease suffix and define the next stable version:

- `1.2.3-4` -> `1.2.4`

### `minor`

Uses the highest Git version as the source.

- `1.2.3` -> `1.3.0`
- `1.2.3-4` -> `1.3.0`

### `major`

Uses the highest Git version as the source.

- `1.2.3` -> `2.0.0`
- `1.2.3-4` -> `2.0.0`

### `prepatch`

First bump to the next stable patch version, then create a numeric prerelease:

- `2.3.3` -> `2.3.4-0`
- `2.3.4-0` -> `2.3.4-1`

### `preminor`

First bump to the next stable minor version, then create a numeric prerelease:

- `2.3.3` -> `2.4.0-0`
- `2.4.0-0` -> `2.4.0-1`

### `premajor`

First bump to the next stable major version, then create a numeric prerelease:

- `2.3.3` -> `3.0.0-0`
- `3.0.0-0` -> `3.0.0-1`

## File Rules

### `package.json`

Update the `version` property.

### `meta.yml`

Schema:

```yml
name: string
version: string
```

If both files exist, both version values must be updated together.

## Git Rules

The tool depends on Git for version operations.

Behavior outside a Git repository:

- version-changing commands warn and do not work
- `resolve` can handle the extended information flow
- `current` still reads JSON files if available

Before creating a new tag, the tool should fetch tags and validate the next possible version.

## Installation

The project should provide:

- `install.sh`
- `install.ps1`
- `Makefile`
