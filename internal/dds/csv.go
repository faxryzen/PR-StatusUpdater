package dds

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

func SaveJSONPullReqsAsCSV(filename string) (error) {
	js, err := os.ReadFile("output/" + filename + ".json")
	if err != nil {
		return fmt.Errorf("cant read file %s; %w", filename, err)
	}

	var prs []PullRequest
	err = json.Unmarshal(js, &prs)
	if err != nil {
		return fmt.Errorf("cant umarshal file %s; %w", filename, err)
	}

	csvFile, err := os.Create("output/" + filename + ".csv")
	if err != nil {
		return fmt.Errorf("cant create csv file: %w", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	header := []string{"Num", "Author", "LabID", "Score", "Debug"}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("cant write csv file: %w", err)
	}

	for _, pr := range prs {
		csvScore := "exist"
		if pr.Score != -1 {
			csvScore = strconv.Itoa(pr.Score)
		}
		row := []string{
			strconv.FormatUint(uint64(pr.Number), 10),
			pr.Author,
			pr.LabID,
			csvScore,
			pr.Debug,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("cant write csv file: %w", err)
		}
	}

	if err := writer.Error(); err != nil {
		return fmt.Errorf("writer error: %w", err)
	}

	return nil
}
