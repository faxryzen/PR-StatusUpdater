package fetcher

import (
	"fmt"
	"os"
	"time"
	"errors"

	"gopkg.in/yaml.v3"
)

const (
	configFilepath = "configs/fetcher/"
)

var ErrEmptyRepos = errors.New("repos.csv is empty")

type GitPlatform interface {
	// Returns pull requests
	//
	// By every query from your cfg (for github) and storage it by name
	// in map
	GetPullRequests(repo Repository) (map[string][]PullRequest, error)
}

type FetchConfig struct {
	Platform string `yaml:"platform"`
	GitHub struct {
		Filter  string            `yaml:"filter"`
		Queries map[string]string `yaml:"queries"`
	} `yaml:"github"`

	//GitLab struct {
	//	ApiURL string `yaml:"api_url"`
	//	Token  string `yaml:"token"`
	//} `yaml:"gitlab"`
}

func Init(cfgname string) (GitPlatform, error) {
	cfg, err := LoadFetchConfig(cfgname)
	if err != nil {
		return nil, fmt.Errorf("init platform error: %w", err)
	}
	switch cfg.Platform {
	case "github":
		return &GitHubPlatform{
			Name: cfg.Platform,
			Filter: cfg.GitHub.Filter,
			Queries: cfg.GitHub.Queries,
			}, nil
	case "gitlab":
		// return &GitLabPlatform{}, nil
		return nil, fmt.Errorf("gitlab is not implemented yet")
	default:
		return nil, fmt.Errorf("unknown platform: %s", cfg.Platform)
	}
}

func LoadFetchConfig(filename string) (*FetchConfig, error) {
	data, err := os.ReadFile(configFilepath + filename)
	if err != nil {
		return nil, fmt.Errorf("fetch cfg read failed: %w", err)
	}

	var cfg FetchConfig

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("fetch cfg parse failed: %w", err)
	}

	return &cfg, nil
}

type Repository struct {
	Name string `yaml:"name"`
	Auth string `yaml:"auth"`
	Gist string `yaml:"gist"`
	Cfg  string `yaml:"conf"`
}

func LoadRepositories() ([]Repository, error) {
	data, err := os.ReadFile(configFilepath + "/repositories.yaml")
	if err != nil {
		return nil, fmt.Errorf("repo cfg read failed: %w", err)
	}

	var cfg []Repository

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("repo cfg parse failed: %w", err)
	}

	return cfg, nil
}

type PullRequest struct {
	Number   uint                 `json:"number"`
	Author   string               `json:"author"`
	Name     string               `json:"name"`
	Times    map[string]time.Time `json:"times"`
	Marks    []string             `json:"marks"`
	Score    float64              `json:"score"`
	Debug    string               `json:"debug"`
}
