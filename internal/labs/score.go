package labs

import (
	"log"
	"strconv"

	f "github.com/faxryzen/pr-updater/internal/fetcher"
)

func DoFormula(score int, do string, howMuch int) int {
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
	dds_acc := cfg.DeadlinesAccept
	dds_red := cfg.DeadlinesReady

	for _, deadline := range dds_acc {
		if pr.Times["created"].After(deadline) {
			pr.Debug += "dd accept expired; "
			score = DoFormula(score, cfg.AcceptDo, cfg.AcceptScore)
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
		mInt, err := strconv.Atoi(mStr)

		if err != nil {
			panic("there's no way")
		}

		pr.Debug += "accept label: " + mStr + "; "
		score += mInt
	}

	if score < 0 {
		score = 0
	}

	pr.Score = score
}
