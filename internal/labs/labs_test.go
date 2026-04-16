package labs

import (
	"testing"
	"time"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

func TestDoFormula(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		do       string
		howMuch  float64
		expected float64
	}{
		{"minus", 10.0, "minus", 2.0, 8.0},
		{"minus negative", 5.0, "minus", 1.5, 3.5},
		{"mult", 10.0, "mult", 2.0, 20.0},
		{"mult zero", 10.0, "mult", 0.0, 0.0},
		{"div", 10.0, "div", 2.0, 5.0},
		{"div by zero", 10.0, "div", 0.0, 10.0},
		{"unknown", 10.0, "plus", 2.0, 10.0},
		{"mult float", 10.0, "mult", 0.5, 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DoFormula(tt.score, tt.do, tt.howMuch)
			if result != tt.expected {
				t.Errorf("DoFormula(%f, %s, %f) = %f, want %f", tt.score, tt.do, tt.howMuch, result, tt.expected)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	baseTime := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	beforeDeadline := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	afterDeadline := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		pr            *f.PullRequest
		cfg           Lab
		expectedScore float64
		expectedDebug string
	}{
		{
			name: "no deadlines, no marks",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": baseTime,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{},
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 10.0,
			expectedDebug: "",
		},
		{
			name: "deadline accept expired",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{beforeDeadline},
				AcceptScore:     2.0,
				AcceptDo:        "minus",
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 8.0,
			expectedDebug: "dd accept expired; ",
		},
		{
			name: "deadline accept with multiply",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{beforeDeadline},
				AcceptScore:     2.0,
				AcceptDo:        "mult",
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 20.0,
			expectedDebug: "dd accept expired; ",
		},
		{
			name: "multiple deadlines accept",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{beforeDeadline, beforeDeadline},
				AcceptScore:     1.0,
				AcceptDo:        "minus",
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 8.0, // two times minus 1
			expectedDebug: "dd accept expired; dd accept expired; ",
		},
		{
			name: "deadline ready expired with fine",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": beforeDeadline,
					"fined":   afterDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{},
				DeadlinesReady:  []time.Time{beforeDeadline},
				ReadyScore:      1.0,
				ReadyDo:         "minus",
			},
			expectedScore: 9.0,
			expectedDebug: "dd fine expired; ",
		},
		{
			name: "deadline ready with merged before fine",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": beforeDeadline,
					"fined":   afterDeadline,
					"merged":  beforeDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{},
				DeadlinesReady:  []time.Time{beforeDeadline},
				ReadyScore:      1.0,
				ReadyDo:         "minus",
			},
			expectedScore: 10.0,
			expectedDebug: "",
		},
		{
			name: "marks processing",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": baseTime,
				},
				Marks: []string{"+2", "-1", "+5"},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{},
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 16.0, // 10 +2 -1 +5 = 16
			expectedDebug: "accept label: +2; accept label: -1; accept label: +5; ",
		},
		{
			name: "combined: deadlines and marks",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
					"fined":   afterDeadline,
				},
				Marks: []string{"+5", "-2"},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{beforeDeadline},
				AcceptScore:     1.0,
				AcceptDo:        "minus",
				DeadlinesReady:  []time.Time{beforeDeadline},
				ReadyScore:      2.0,
				ReadyDo:         "minus",
			},
			expectedScore: 10.0, // 10 -1 -2 +5 -2 = 10
			expectedDebug: "dd accept expired; dd fine expired; accept label: +5; accept label: -2; ",
		},
		{
			name: "score cannot be negative",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
				},
				Marks: []string{},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       1.0,
				DeadlinesAccept: []time.Time{beforeDeadline},
				AcceptScore:     2.0,
				AcceptDo:        "minus",
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 0.0,
			expectedDebug: "dd accept expired; ",
		},
		{
			name: "mult float with marks",
			pr: &f.PullRequest{
				Times: map[string]time.Time{
					"created": afterDeadline,
				},
				Marks: []string{"+0.5"},
				Debug: "",
			},
			cfg: Lab{
				BaseScore:       10.0,
				DeadlinesAccept: []time.Time{beforeDeadline},
				AcceptScore:     0.5,
				AcceptDo:        "mult",
				DeadlinesReady:  []time.Time{},
			},
			expectedScore: 5.5,
			expectedDebug: "dd accept expired; accept label: +0.5; ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CalculateScore(tt.pr, tt.cfg)

			if tt.pr.Score != tt.expectedScore {
				t.Errorf("Score = %f, want %f", tt.pr.Score, tt.expectedScore)
			}

			if tt.pr.Debug != tt.expectedDebug {
				t.Errorf("Debug = %q, want %q", tt.pr.Debug, tt.expectedDebug)
			}
		})
	}
}

func TestNoFineLabel(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	pr := &f.PullRequest{
		Times: map[string]time.Time{
			"created": now,
		},
		Marks: []string{},
		Debug: "",
	}

	cfg := Lab{
		BaseScore:      10.0,
		DeadlinesReady: []time.Time{yesterday},
		ReadyScore:     1.0,
		ReadyDo:        "minus",
	}

	CalculateScore(pr, cfg)

	if pr.Score != 10.0 {
		t.Errorf("Expected score 10.0, got %f", pr.Score)
	}

	if contains(pr.Debug, "dd fine expired") {
		t.Error("Debug should not contain 'dd fine expired' when no fine label")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr) && containsHelper(s, substr))))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
