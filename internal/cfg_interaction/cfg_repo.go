package cfginteraction

import (
	"encoding/csv"
	"fmt"
	"os"
)

const configFilepath = "configs/"

func CfgReadRepoCSV() ([][]string, error) {
	file, err := os.Open(configFilepath + "repos.csv")

	if err != nil {
		file, err = os.Create(configFilepath + "repos.csv")
		if err != nil {
			return nil, fmt.Errorf("failed to create cfg: %w", err)
		}
		defer file.Close()

		return [][]string{}, nil
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("invalid cfg: %w", err)
	}

	return records, nil
}

func CfgWriteRepoCSV(record []string) (error) {
	file, err := os.Open(configFilepath + "repos.csv")

	if err != nil {
		return fmt.Errorf("failed open cfg: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	
	err = writer.Write(record)

	if err != nil {
		return fmt.Errorf("failed write into cfg: %w", err)
	}
	return nil
}