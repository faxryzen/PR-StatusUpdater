package dds

import (
	"strconv"
)

func CalculateScore(pr *PullRequest, cfg Lab) {
	score := cfg.BaseScore
	dds_acc := cfg.DeadlinesAccept
	dds_red := cfg.DeadlinesReady

	for _, deadline := range dds_acc {
		if pr.Times["created"].After(deadline) {
			pr.Debug += "dd accept proeban; "
			score--
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
			pr.Debug += "dd fine proeban; "
			score--
		}
	}

	marks := pr.Marks

	for _, mStr := range marks {
		mInt, err := strconv.Atoi(mStr)

		if err != nil {
			panic("there's no fucking way")
		}

		pr.Debug += "accept label: " + mStr + "; "
		score += mInt
	}

	if score < 0 {
		score = 0
	}

	pr.Score = score
}
