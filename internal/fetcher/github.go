package fetcher

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	filterFilepath = "configs/fetcher/filters/"
	queryFilepath  = "configs/fetcher/queries/"
)

type GitHubPlatform struct {
	Name    string
	Filter  string
	Queries map[string]string
}

type Query struct {
	Type string
	Data string
}

func (g *GitHubPlatform) GetQueries(repo Repository) ([]Query, error) {
	res := []Query{}

	for name, filename := range g.Queries {

		rawgraph, err := os.ReadFile(queryFilepath + filename)
		if err != nil {
			return nil, err
		}

		data := strings.ReplaceAll(string(rawgraph), "$owner", repo.Auth)
		data = strings.ReplaceAll(data, "$name", repo.Name)

		res = append(res, Query{
			Type: name,
			Data: data,
		})
	}

	return res, nil
}

func (g *GitHubPlatform) GetPullRequests(repo Repository) (map[string][]PullRequest, error) {
	jqFilter, err := os.ReadFile(filterFilepath + g.Filter)
	if err != nil {
		return nil, fmt.Errorf("getpullreq failed: %w", err)
	}

	querys, err := g.GetQueries(repo)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]PullRequest)

	for _, q := range querys {

		pullRequests, err := g.execFetch(q.Data, string(jqFilter))
		if err != nil {
			return nil, err
		}

		result[q.Type] = pullRequests
	}
	return result, nil
}

func (g *GitHubPlatform) execFetch(query string, jqFilter string) ([]PullRequest, error) {
	cmd := exec.Command(
		"gh", "api", "graphql",
		"-f", "query="+query,
		"--jq", jqFilter)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}

	return parsePullRequests(out), nil
}


func /*(g *GitHubPlatform)*/ UploadCSVtoGist(repo Repository) (error) {

	err := SaveJSONPullReqsAsCSV(repo.Name)
	if err != nil {
		return err
	}

	cmd := exec.Command("gh", "gist", "edit", repo.Gist, "output/" + repo.Name + ".csv")

	_, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
