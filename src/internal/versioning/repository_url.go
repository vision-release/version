package versioning

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrNoRemoteURL            = errors.New("no git remote url found")
	ErrUnsupportedGitProvider = errors.New("unsupported git provider")
)

type RepositoryURLs struct {
	GitURL      string
	Repository  string
	PipelineURL string
}

func (a *App) RepositoryURLs() (RepositoryURLs, error) {
	if !a.git.IsRepo(a.root) {
		return RepositoryURLs{}, ErrNotGitRepo
	}

	remoteURL, err := a.git.RemoteURL(a.root, "origin")
	if err != nil {
		return RepositoryURLs{}, err
	}
	if remoteURL == "" {
		return RepositoryURLs{}, ErrNoRemoteURL
	}

	return ParseRepositoryURLs(remoteURL)
}

func ParseRepositoryURLs(remoteURL string) (RepositoryURLs, error) {
	host, path, err := parseGitRemote(remoteURL)
	if err != nil {
		return RepositoryURLs{}, err
	}

	publicURL := "https://" + host + "/" + path

	switch providerForHost(host) {
	case "github":
		return RepositoryURLs{
			GitURL:      remoteURL,
			Repository:  publicURL,
			PipelineURL: publicURL + "/actions",
		}, nil
	case "gitlab":
		return RepositoryURLs{
			GitURL:      remoteURL,
			Repository:  publicURL,
			PipelineURL: publicURL + "/-/pipelines",
		}, nil
	default:
		return RepositoryURLs{}, fmt.Errorf("%w: %s", ErrUnsupportedGitProvider, host)
	}
}

func parseGitRemote(remoteURL string) (host, path string, err error) {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return "", "", ErrNoRemoteURL
	}

	if strings.Contains(remoteURL, "://") {
		parsed, parseErr := url.Parse(remoteURL)
		if parseErr != nil {
			return "", "", parseErr
		}
		host = parsed.Hostname()
		path = strings.TrimPrefix(parsed.Path, "/")
	} else {
		at := strings.Index(remoteURL, "@")
		colon := strings.Index(remoteURL, ":")
		if at == -1 || colon == -1 || colon < at {
			return "", "", fmt.Errorf("unsupported git remote url: %s", remoteURL)
		}
		host = remoteURL[at+1 : colon]
		path = remoteURL[colon+1:]
	}

	path = strings.TrimSuffix(path, ".git")
	path = strings.Trim(path, "/")
	if host == "" || path == "" {
		return "", "", fmt.Errorf("unsupported git remote url: %s", remoteURL)
	}

	return host, path, nil
}

func providerForHost(host string) string {
	lowerHost := strings.ToLower(host)
	switch {
	case strings.Contains(lowerHost, "github"):
		return "github"
	case strings.Contains(lowerHost, "gitlab"):
		return "gitlab"
	default:
		return ""
	}
}
