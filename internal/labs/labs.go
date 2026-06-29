package labs

import (
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

const (
	configFilepath = "configs/labs/"
)

type LabsConfig struct {
	Labs []Lab `json:"labs" yaml:"labs"`
}

type Lab struct {
	ID              string      `json:"id"                 yaml:"id"`
	BaseScore       float64     `json:"base_score"         yaml:"base_score"`

	DeadlinesAccept []time.Time `json:"dds_acceptance"     yaml:"dds_acceptance"`
	AcceptScore     float64     `json:"acceptance_score"   yaml:"acceptance_score"`
	AcceptDo        string      `json:"acceptance_do"      yaml:"acceptance_do"`

	DeadlinesReady  []time.Time `json:"dds_readiness"      yaml:"dds_readiness"`
	ReadyScore      float64     `json:"readiness_score"    yaml:"readiness_score"`
	ReadyDo         string      `json:"readiness_do"       yaml:"readiness_do"`
}

// Returns Lab by name (from labs-<cfg-fetcher-name>.yaml)
func LoadDeadlines(cfgFetcher string) (map[string]Lab, error) {
	data, err := os.ReadFile(configFilepath + "labs-" + cfgFetcher)
	if err != nil {
		return nil, err
	}

	var cfg LabsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	labs := make(map[string]Lab)
	for _, lab := range cfg.Labs {
		applyDefaults(&lab)
		labs[lab.ID] = lab
	}

	return labs, nil
}

func applyDefaults(lab *Lab) {
	if len(lab.DeadlinesAccept) > 0 {
		if lab.AcceptScore == 0 {
			lab.AcceptScore = 1
		}
		if lab.AcceptDo == "" {
			lab.AcceptDo = "minus"
		}
	}

	if len(lab.DeadlinesReady) > 0 {
		if lab.ReadyScore == 0 {
			lab.ReadyScore = 1
		}
		if lab.ReadyDo == "" {
			lab.ReadyDo = "minus"
		}
	}
}

// Removes untitled pull requests - they didn't pass jq-filter and likely aren't labs.
//
// Converts Time.time objects to Moscow and removes time from PullRequest.Times
// if it hasn't been fetched (this didn't happen)
func ConvertToLabs(raw []f.PullRequest) ([]f.PullRequest) {
	clean := []f.PullRequest{}
	loc, _ := time.LoadLocation("Europe/Moscow")

	for _, pr := range raw {

		if pr.Name == "" {
			continue
		}

		pr.Name = strings.ToUpper(pr.Name)

		pr.Times["created"] = pr.Times["created"].In(loc)

		if pr.Times["fined"].IsZero() {
			delete(pr.Times, "fined")
		} else {
			pr.Times["fined"] = pr.Times["fined"].In(loc)
		}
		if pr.Times["merged"].IsZero() {
			delete(pr.Times, "merged")
		} else {
			pr.Times["merged"] = pr.Times["merged"].In(loc)
		}

		pr.Score = -1

		clean = append(clean, pr)
	}
	return clean
}