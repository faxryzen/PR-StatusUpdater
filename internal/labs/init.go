package labs

import (
	"encoding/json"
	"strings"
	"time"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

const (
	configFilepath = "configs/"
)

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

func FetchLabsToJSON(repo f.Repository) ([]byte, error) {
	labs, err := LoadDeadlines()
	if err != nil {
		return nil, err
	}

	git, err := f.Init("github-volgarenok.yaml")
	if err != nil {
		return nil, err
	}

	prs, err := git.GetPullRequests(repo)
	if err != nil {
		return nil, err
	}

	mergedPRs := ConvertToLabs(prs["merged"])
	for i := range mergedPRs {
		CalculateScore(&mergedPRs[i], labs[mergedPRs[i].Name])
	}

	j, err := json.MarshalIndent(append(mergedPRs, ConvertToLabs(prs["open"])...), "", " ")
	if err != nil {
		return nil, err
	}

	return j, nil
}
