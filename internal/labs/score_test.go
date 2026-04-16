package labs

import (
	"encoding/json"
	"testing"
	"time"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

func TestCalculateScoreOnMergedPRs(t *testing.T) {
	labs := map[string]Lab{
		"LAB1": {
			ID:        "LAB1",
			BaseScore: 10.0,
			DeadlinesAccept: []time.Time{
				time.Date(2025, 1, 5, 23, 59, 0, 0, time.UTC),
			},
			AcceptScore: 1.0,
			AcceptDo:    "minus",
			DeadlinesReady: []time.Time{
				time.Date(2025, 1, 10, 23, 59, 0, 0, time.UTC),
			},
			ReadyScore: 2.0,
			ReadyDo:    "minus",
		},
		"LAB2": {
			ID:        "LAB2",
			BaseScore: 20.0,
			DeadlinesAccept: []time.Time{
				time.Date(2025, 2, 1, 23, 59, 0, 0, time.UTC),
			},
			AcceptScore: 1.5,
			AcceptDo:    "minus",
			DeadlinesReady: []time.Time{
				time.Date(2025, 2, 10, 23, 59, 0, 0, time.UTC),
			},
			ReadyScore: 3.0,
			ReadyDo:    "mult",
		},
	}

	mergedPRs := []f.PullRequest{
		{
			Number: 101,
			Name:   "LAB1",
			Author: "student1",
			Times: map[string]time.Time{
				"created": time.Date(2025, 1, 6, 12, 0, 0, 0, time.UTC), // dead accept
				"fined":   time.Date(2025, 1, 9, 12, 0, 0, 0, time.UTC), // got ready
			},
			Marks: []string{"+2", "-1"},
			Score: -1,
		},
		{
			Number: 102,
			Name:   "LAB1",
			Author: "student2",
			Times: map[string]time.Time{
				"created": time.Date(2025, 1, 3, 12, 0, 0, 0, time.UTC),  // got accept
				"fined":   time.Date(2025, 1, 12, 12, 0, 0, 0, time.UTC), // dead ready
			},
			Marks: []string{"+5"},
			Score: -1,
		},
		{
			Number: 103,
			Name:   "LAB2",
			Author: "student3",
			Times: map[string]time.Time{
				"created": time.Date(2025, 2, 2, 12, 0, 0, 0, time.UTC), // dead accept
				"fined":   time.Date(2025, 2, 5, 12, 0, 0, 0, time.UTC), // got ready
			},
			Marks: []string{"-2"},
			Score: -1,
		},
	}

	testPRs := make([]f.PullRequest, len(mergedPRs))
	copy(testPRs, mergedPRs)

	for i := range testPRs {
		lab, ok := labs[testPRs[i].Name]
		if !ok {
			t.Fatalf("Lab %s not found", testPRs[i].Name)
		}
		CalculateScore(&testPRs[i], lab)
	}

	expectedScores := []float64{
		10.0 - 1.0 + 2.0 - 1.0, // LAB1: base 10 -1 (accept) +2 -1 = 10
		10.0 - 2.0 + 5.0,       // LAB1: base 10 -2 (ready) +5 = 13
		20.0 - 1.5 - 2.0,       // LAB2: base 20 -1.5 (accept) -2 = 16.5
	}

	for i, pr := range testPRs {
		if pr.Score != expectedScores[i] {
			t.Errorf("PR #%d (%s): expected score %.2f, got %.2f",
				pr.Number, pr.Name, expectedScores[i], pr.Score)
		}
	}

	jsonData, err := json.MarshalIndent(testPRs, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON output is empty")
	}

	var parsed []f.PullRequest
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if len(parsed) != len(testPRs) {
		t.Errorf("JSON has %d items, expected %d", len(parsed), len(testPRs))
	}
}
