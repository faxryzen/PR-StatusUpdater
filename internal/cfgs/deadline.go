package cfgs

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	cfgTimeLayout = "02.01.2006"
)

type DeadlinesConfig struct {
	Labs      []Lab `json:"labs"`
}

type Lab struct {
	ID        string      `json:"id"`
	BaseScore int         `json:"base_score"`
	Deadlines []time.Time `json:"deadlines"`
}

func ToMoscow(t time.Time) time.Time {
	loc, _ := time.LoadLocation("Europe/Moscow")
	return t.In(loc)
}

func stringToMoscowTime(oldTime string) (time.Time, error) {
	newTime, err := time.Parse(time.RFC3339, oldTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("ERROR: failed to parse time: %w", err)
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic("location invalid")
	}
	newTime = newTime.In(loc)
	return newTime, nil
}

func LoadDeadlines() (*DeadlinesConfig, map[string]Lab, error) {
	data, err := os.ReadFile("deadlines.json")
	if err != nil {
		return nil, nil, err
	}

	var cfg DeadlinesConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, nil, err
	}

	labs := make(map[string]Lab)
	for _, lab := range cfg.Labs {
		labs[lab.ID] = lab
	}

	return &cfg, labs, nil
}

func NormalizeDeadlines(labs map[string]Lab) {
	loc, _ := time.LoadLocation("Europe/Moscow")

	for id, lab := range labs {
		for i, d := range lab.Deadlines {
			lab.Deadlines[i] = time.Date(
				d.Year(), d.Month(), d.Day(),
				d.Hour(), d.Minute(), 0, 0,
				loc,
			)
		}
		labs[id] = lab
	}
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

	for _, deadline := range lab.Deadlines {
		if mergedAt.After(deadline) {
			score--
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

