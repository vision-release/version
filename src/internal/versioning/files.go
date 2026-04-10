package versioning

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrNoVersionFile = errors.New("no version file found")

type VersionFiles struct {
	PackageJSON *JSONFile
	MetaYML     *YAMLFile
}

type JSONFile struct {
	Path    string
	Version Version
	Raw     map[string]any
}

type YAMLFile struct {
	Path       string
	Name       string
	Version    Version
	TagPattern string
}

func InitMetaYML(root, name string, version Version) (YAMLFile, error) {
	file := YAMLFile{
		Path:       filepath.Join(root, "meta.yml"),
		Name:       strings.TrimSpace(name),
		Version:    version,
		TagPattern: DefaultTagPattern,
	}

	if file.Name == "" {
		return YAMLFile{}, fmt.Errorf("project name cannot be empty")
	}

	if err := writeYAMLFile(file); err != nil {
		return YAMLFile{}, err
	}

	return file, nil
}

func LoadVersionFiles(root string) (VersionFiles, error) {
	files := VersionFiles{}

	packagePath := filepath.Join(root, "package.json")
	if exists(packagePath) {
		file, err := loadJSONVersionFile(packagePath)
		if err != nil {
			return VersionFiles{}, err
		}

		files.PackageJSON = &file
	}

	metaPath := filepath.Join(root, "meta.yml")
	if exists(metaPath) {
		file, err := loadYAMLVersionFile(metaPath)
		if err != nil {
			return VersionFiles{}, err
		}

		files.MetaYML = &file
	}

	if files.PackageJSON == nil && files.MetaYML == nil {
		return VersionFiles{}, ErrNoVersionFile
	}

	return files, nil
}

func (f VersionFiles) PrimaryVersion() string {
	if f.PackageJSON != nil {
		return f.PackageJSON.Version.String()
	}

	return f.MetaYML.Version.String()
}

func (f VersionFiles) TagPattern() string {
	if f.MetaYML != nil && f.MetaYML.TagPattern != "" {
		return f.MetaYML.TagPattern
	}
	return DefaultTagPattern
}

func (f VersionFiles) PrimaryParsedVersion() *Version {
	if f.PackageJSON != nil {
		version := f.PackageJSON.Version
		return &version
	}

	if f.MetaYML != nil {
		version := f.MetaYML.Version
		return &version
	}

	return nil
}

func (f VersionFiles) Sync(version Version) error {
	if f.PackageJSON != nil {
		f.PackageJSON.Raw["version"] = version.String()
		if err := writeJSONFile(f.PackageJSON.Path, f.PackageJSON.Raw); err != nil {
			return err
		}
	}

	if f.MetaYML != nil {
		f.MetaYML.Version = version
		if err := writeYAMLFile(*f.MetaYML); err != nil {
			return err
		}
	}

	return nil
}

func loadJSONVersionFile(path string) (JSONFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return JSONFile{}, err
	}

	var raw map[string]any
	if err := json.Unmarshal(content, &raw); err != nil {
		return JSONFile{}, fmt.Errorf("invalid json in %s: %w", path, err)
	}

	value, ok := raw["version"].(string)
	if !ok || value == "" {
		return JSONFile{}, fmt.Errorf("%s does not contain a string version", path)
	}

	version, err := ParseVersion(value)
	if err != nil {
		return JSONFile{}, fmt.Errorf("%s version is invalid: %w", path, err)
	}

	return JSONFile{
		Path:    path,
		Version: version,
		Raw:     raw,
	}, nil
}

func loadYAMLVersionFile(path string) (YAMLFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return YAMLFile{}, err
	}

	data := make(map[string]string)
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return YAMLFile{}, fmt.Errorf("invalid yml in %s", path)
		}

		data[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"'")
	}

	value := data["version"]
	if value == "" {
		return YAMLFile{}, fmt.Errorf("%s does not contain a version", path)
	}

	version, err := ParseVersion(value)
	if err != nil {
		return YAMLFile{}, fmt.Errorf("%s version is invalid: %w", path, err)
	}

	return YAMLFile{
		Path:       path,
		Name:       data["name"],
		Version:    version,
		TagPattern: data["tag_pattern"],
	}, nil
}

func writeJSONFile(path string, content map[string]any) error {
	data, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeYAMLFile(file YAMLFile) error {
	lines := []string{}
	if file.Name != "" {
		lines = append(lines, fmt.Sprintf("name: %s", file.Name))
	}
	lines = append(lines, fmt.Sprintf("version: %s", file.Version.String()))
	if file.TagPattern != "" {
		lines = append(lines, fmt.Sprintf("tag_pattern: %s", file.TagPattern))
	}
	lines = append(lines, "")

	return os.WriteFile(file.Path, []byte(strings.Join(lines, "\n")), 0o644)
}

func exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
