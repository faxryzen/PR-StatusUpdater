package labs

import (
	"encoding/json"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

func FetchLabsToJSON(repo f.Repository) ([]byte, error) {
	labs, err := LoadDeadlines(repo.Cfg)
	if err != nil {
		return nil, err
	}

	git, err := f.Init(repo.Cfg)
	if err != nil {
		return nil, err
	}

	prs, err := git.GetPullRequests(repo)
	if err != nil {
		return nil, err
	}

	mergedPRs := ConvertToLabs(prs["merged"])
	openPRs   := ConvertToLabs(prs["open"])

	mergedPRsFixed := []f.PullRequest{}
	openPRsFixed :=   []f.PullRequest{}

	for i := range mergedPRs {
		lab_cfg, exist := labs[openPRs[i].Name]
		if exist {
			CalculateScore(&mergedPRs[i], lab_cfg)
			mergedPRsFixed = append(mergedPRsFixed, mergedPRs[i])
		}
	}

	for i := range openPRs {
		lab_cfg, exist := labs[openPRs[i].Name]
		if exist {
			CalculateScore(&openPRs[i], lab_cfg)
			openPRsFixed = append(openPRsFixed, openPRs[i])
		}
	}

	j, err := json.MarshalIndent(append(mergedPRsFixed, openPRsFixed...), "", " ")
	if err != nil {
		return nil, err
	}

	return j, nil
}
