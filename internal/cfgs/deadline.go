package cfgs

import (
	"time"
	"strings"
	"fmt"
	"strconv"
)

const (
	oldTimeLayout = "dd-mm-yyyy 15:04:05"
	cfgTimeLayout = "dd.mm.yyyy"
	newTimeLayout = "dd.mm.yyyy 15:04:05"
)

func stringToMoscowTime(oldTime string) (time.Time, error) {
	oldTime = strings.ReplaceAll(oldTime, "T", " ")
	oldTime = strings.ReplaceAll(oldTime, "Z", "")

	newTime, err := time.Parse(oldTimeLayout, oldTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("ERROR: failed to parse time: %w", err)
	}

	return newTime, nil
}

func getDeadlines() (map[string][]time.Time, error) {
	records, err := ScanRecords("deadlines.csv")
	if err != nil {
		return nil, err
	}

	dds := make(map[string][]time.Time)

	for _, line := range records {
		temp := []time.Time{}
		for i := 1; i < len(line); i++ {
			t, err := time.Parse(cfgTimeLayout, line[i])
			if err != nil {
				panic("cfg dds wrong")
			}
			temp = append(temp, t)
		}
		dds[strings.ToUpper(line[0])] = temp
	}
	return dds, nil
}

//Считает штрафы на основе дедлайнов и конфига deadlines.csv
func TransformPenaltyData(raw string, iLab uint, iTimes uint) (string, error) {
	dds, err := getDeadlines()
	if err != nil {
		panic("omg")
	}

	var data string

	lines := strings.Split(string(raw), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		
		pen, err := calcPenalty(dds[parts[iLab]], parts[iTimes:iTimes + 1])
		if err != nil {
			panic("failed to calc penalty")
		}
		for i := 0; i < len(parts) - 2; i++ {
			data += parts[i] + ";"
		}
		data += strconv.Itoa(pen) + "\n"
	}

	return data, nil
}

func calcPenalty(dds []time.Time, ats []string) (int, error) {
	//todo
	return  0, nil
}