package dds

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

const (
	configFilepath = "configs/"
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

type PullRequest struct {
	Number  uint                 `json:"number"`
	Author  string               `json:"author"`
	LabID   string               `json:"labID"`
	Times   map[string]time.Time `json:"times"` //created fined merged
	Marks   []string             `json:"marks"`
	Score   int                  `json:"score"`
	Debug   string               `json:"debug"`
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

	return labs, nil
}

func CalculateScore(pr *PullRequest, cfg Lab) {
	score := cfg.BaseScore
	dds_acc := cfg.DeadlinesAccept
	dds_red := cfg.DeadlinesReady

	for _, deadline := range dds_acc {
		if pr.Times["created"].After(deadline) {
			pr.Debug += "dd accept proeban; "
			score--
		}
	}

	fineOrMergeTime := pr.Times["fined"]

	if _, ok := pr.Times["merged"]; ok {
		if pr.Times["merged"].Before(pr.Times["fined"]) {

			fineOrMergeTime = pr.Times["merged"]
		}
	}
	for _, deadline := range dds_red {
		if fineOrMergeTime.After(deadline) {
			pr.Debug += "dd fine proeban; "
			score--
		}
	}

	marks := pr.Marks

	for _, mStr := range marks {
		mInt, err := strconv.Atoi(mStr)

		if err != nil {
			panic("there's no fucking way")
		}

		pr.Debug += "accept label: " + mStr + "; "
		score += mInt
	}

	if score < 0 {
		score = 0
	}

	pr.Score = score
}
