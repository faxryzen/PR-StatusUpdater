package labs

import (
	"log"
	"strconv"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

func DoFormula(score float64, do string, howMuch float64) float64 {
	switch do {
	case "minus":
		score -= howMuch
	case "mult":
		score *= howMuch
	case "div":
		if howMuch != 0 {
			score /= howMuch
		}
	default:
		log.Println("Unknown do: " + do)
	}
	return score
}

func CalculateScore(pr *f.PullRequest, cfg Lab) {
	score := cfg.BaseScore
	dds_crt := cfg.DeadlinesCreate
	dds_red := cfg.DeadlinesReady

	for _, deadline := range dds_crt {
		if pr.Times["created"].After(deadline) {
			pr.Debug += "dd creation expired; "
			score = DoFormula(score, cfg.CreateDo, cfg.CreateScore)
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
			pr.Debug += "dd fine expired; "
			score = DoFormula(score, cfg.ReadyDo, cfg.ReadyScore)
		}
	}

	marks := pr.Marks

	for _, mStr := range marks {
		mFloat, err := strconv.ParseFloat(mStr, 64)

		if err != nil {
			pr.Debug += "invalid mark: " + mStr + "; "
			continue
	}

		pr.Debug += "accept label: " + mStr + "; "
		score += mFloat
	}

	if score < 0 {
		score = 0
	}

	pr.Score = score
}
