package dds

import (
	"encoding/json"
)

const (
	configFilepath = "configs/"
	filterFilepath = "internal/dds/filters/"
	queryFilepath = "internal/dds/queries/"
)

func UnloadLabs(repo Repository) ([]byte, error) {
	labs, err := LoadDeadlines()
	if err != nil {
		return nil, err
	}

	querys, err := GetGraphQLForGit(repo)
	if err != nil {
		return nil, err
	}

	mergedPRs, err := GetPullRequests(querys.MergedPRsQuery)
	if err != nil {
		return nil, err
	}

	openPRs, err := GetPullRequests(querys.OpenPRsQuery)
	if err != nil {
		return nil, err
	}

	for i := range mergedPRs {
		CalculateScore(&mergedPRs[i], labs[mergedPRs[i].LabID])
	}

	j, err := json.MarshalIndent(append(mergedPRs, openPRs...), "", " ")
	if err != nil {
		return nil, err
	}

	return j, nil
}
