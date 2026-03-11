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

	for i := range mergedPRs {
		CalculateScore(&mergedPRs[i], labs[mergedPRs[i].Name])
	}

	j, err := json.MarshalIndent(append(mergedPRs, openPRs...), "", " ")
	if err != nil {
		return nil, err
	}

	return j, nil
}
