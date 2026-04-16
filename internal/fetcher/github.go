package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func (g *GitHubPlatform) GetQueries() ([]Query, error) {
	res := []Query{}

	for name, filename := range g.Queries {
		rawgraph, err := os.ReadFile(queryFilepath + filename)
		if err != nil {
			return nil, err
		}

		res = append(res, Query{
			Type: name,
			Data: string(rawgraph),
		})
	}

	return res, nil
}

func (g *GitHubPlatform) GetPullRequests(repo Repository) (map[string][]PullRequest, error) {
	jq, err := os.ReadFile(filterFilepath + g.Filter)
	if err != nil {
		return nil, fmt.Errorf("getpullreq failed: %w", err)
	}

	querys, err := g.GetQueries()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]PullRequest)

	for _, q := range querys {
		var pullRequests []PullRequest
		var cursor string
		log.Println("Fetching prs for " + repo.Auth + "/" + repo.Name + " by query: " + q.Type)
		for {
			raw, err := g.execFetch(q.Data, string(jq), repo.Auth, repo.Name, cursor)
			log.Println("+ fetched prs")
			if err != nil {
				return nil, err
			}
			prs, newcursor, hasNext, err := parseResponse(raw)
			if err != nil {
				return nil, err
			}
			pullRequests = append(pullRequests, prs...)
			if !hasNext {
				break
			}
			cursor = newcursor
		}
		log.Println("Done fetching by query: " + q.Type)
		result[q.Type] = pullRequests
	}
	return result, nil
}

func (g *GitHubPlatform) execFetch(query string, jq string, owner, name, cursor string) ([]byte, error) {
	vars := map[string]any{
		"owner": owner,
		"name":  name,
	}
	if cursor != "" {
		vars["cursor"] = cursor
	}

	payload := map[string]any{
		"query":     query,
		"variables": vars,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("gh", "api", "graphql", "--jq", jq, "--input", "-")
	cmd.Stdin = bytes.NewReader(payloadJSON)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w, output: %s", err, out)
	}

	if len(out) == 0 {
		return []byte("{}"), nil
	}

	return out, nil
}

func parseResponse(execOut []byte) ([]PullRequest, string, bool, error) {
	var result struct {
		PRs         []PullRequest `json:"prs"`
		HasNextPage bool          `json:"hasNextPage"`
		EndCursor   string        `json:"endCursor"`
	}

	err := json.Unmarshal(execOut, &result)
	if err != nil {
		return nil, "", false, fmt.Errorf("parse response failed: %w", err)
	}

	return result.PRs, result.EndCursor, result.HasNextPage, nil
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
