package cfgs

import (
	"encoding/csv"
	"fmt"
	"os"
	"github.com/faxryzen/pr-updater/internal/fmtc"
)

const configFilepath = "configs/"

//Открывает repos.csv из папки конфигов, возвращает записи
func scanRecords() ([][]string, error) {
	file, err := os.Open(configFilepath + "repos.csv")
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

func writeRecords(records [][]string) error {
	file, err := os.Create(configFilepath + "repos.csv")
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

//Функция с консольным взаимодействием,
//возвращает два элемента, первый - репо, второй - автор
func GetRepo() ([]string, error) {
	records, err := scanRecords()
	if err != nil {
		return nil, err
	}

	fmtc.Cyan.Println("You have next options:")
	fmtc.Yellow.Println("0: Add new existing repo")

	repoCounter := uint(len(records))

	if repoCounter != 0 {
		fmtc.Green.Println("OR")
		fmtc.Cyan.Println("You can use your recent repos:")

		for i, str := range records {
			fmtc.Yellow.Printf("%d: %s by %s\n", i + 1, str[0], str[1])
		}
	}

	var inputRepo uint

	for {
		fmtc.Green.Println("Type the number of option to continue")
		fmt.Print("> ")
		_, err := fmt.Scanln(&inputRepo)
		if err == nil {
			if inputRepo > repoCounter {
				fmtc.Red.Println("Invalid number, please try again")
				continue
			}
			break
		}

		fmtc.Red.Println("Input error, please try again")
		var discard string
		_, err = fmt.Scanln(&discard)
		if err != nil {
			return nil, fmt.Errorf("ERROR: failed to restart scan %w", err)
		}
	}

	if inputRepo == 0 {
		fmtc.Cyan.Println("Type repo name and owner (use space for separator)")
		fmt.Print("> ")
		var (
			name  string
			owner string
		)
		_, err := fmt.Scanln(&name, &owner)
		if err != nil {
			return nil, fmt.Errorf("ERROR: failed to scan repo, owner %w", err)
		}
		record := []string{name, owner}
		records = append([][]string{record}, records...)
	} else {
		i := inputRepo - 1
		record := []string{records[i][0], records[i][1]}
		records = append(records[:i], records[i + 1:]...)
		records = append([][]string{record}, records...)
	}

	if err := writeRecords(records); err != nil {
		return nil, err
	}

	return records[0], nil
}
