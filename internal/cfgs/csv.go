package cfgs

import (
	"encoding/csv"
	"fmt"
	"os"
)

const configFilepath = "configs/"

//Открывает файл (например repos.csv) из папки конфигов, возвращает записи
func ScanRecords(filename string) ([][]string, error) {
	file, err := os.Open(configFilepath + filename)
	if err != nil {
		if os.IsNotExist(err) {
			return [][]string{}, nil
		}
		return nil, fmt.Errorf("ERROR: failed to open cfg: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ERROR: invalid cfg: %w", err)
	}

	return records, nil
}

func WriteRecords(records [][]string, filename string) error {
	file, err := os.Create(configFilepath + filename)
	if err != nil {
		return fmt.Errorf("ERROR: failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("ERROR: failed to write records: %w", err)
	}

	writer.Flush()
	return writer.Error()
}