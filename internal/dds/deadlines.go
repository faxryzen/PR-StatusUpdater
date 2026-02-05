package dds

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	configFilepath = "configs/"
)

type DeadlinesConfig struct {
	Labs      []Lab `json:"labs"`
}

type Lab struct {
	ID              string      `json:"id"`
	BaseScore       int         `json:"base_score"`
	DeadlinesAccept []time.Time `json:"dds_acceptance"`
	DeadlinesReady  []time.Time `json:"dds_readiness"`
}

func ToMoscow(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Europe/Moscow")
	return t.In(loc)
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

	fmt.Println("===============================labs:")
	fmt.Println(labs)

	return labs, nil
}

func ExtractLabID(title string) string {
	parts := strings.Split(title, "/")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func MatchLab(title string, labs map[string]Lab) (Lab, bool) {
	id := ExtractLabID(title)
	lab, ok := labs[id]
	return lab, ok
}

func CalculateScore(lab Lab, mergedAt time.Time) int {
	score := lab.BaseScore
	/*
	for _, deadline := range lab.Deadlines {
		if mergedAt.After(deadline) {
			score--
		}
	}
*/
	if score < 0 {
		score = 0
	}

	return score
}

