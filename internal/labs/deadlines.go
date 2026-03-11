package labs

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LabsConfig struct {
	Labs []Lab `json:"labs" yaml:"labs"`
}

type Lab struct {
	ID              string      `json:"id"             yaml:"id"`
	BaseScore       int         `json:"base_score"     yaml:"base_score"`
	DeadlinesAccept []time.Time `json:"dds_acceptance" yaml:"dds_acceptance"`
	DeadlinesReady  []time.Time `json:"dds_readiness"  yaml:"dds_readiness"`
}

func LoadDeadlines() (map[string]Lab, error) {
	data, err := os.ReadFile(configFilepath + "labs.yaml")
	if err != nil {
		return nil, err
	}

	var cfg LabsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	labs := make(map[string]Lab)
	for _, lab := range cfg.Labs {
		labs[lab.ID] = lab
	}

	return labs, nil
}
