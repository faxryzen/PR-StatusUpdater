package dds

import (
	"encoding/json"
	"os"
	"time"
)

type DeadlinesConfig struct {
	Labs []Lab `json:"labs"`
}

type Lab struct {
	ID              string      `json:"id"`
	BaseScore       int         `json:"base_score"`
	DeadlinesAccept []time.Time `json:"dds_acceptance"`
	DeadlinesReady  []time.Time `json:"dds_readiness"`
}

func LoadDeadlines() (map[string]Lab, error) {
	data, err := os.ReadFile(configFilepath + "deadlines.json")
	if err != nil {
		return nil, err
	}

	var cfg DeadlinesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	labs := make(map[string]Lab)
	for _, lab := range cfg.Labs {
		labs[lab.ID] = lab
	}

	return labs, nil
}
