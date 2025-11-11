package cfgs

import (
	"time"
	"strings"
	"fmt"
	"strconv"
)

const (
	cfgTimeLayout = "02.01.2006"
	//newTimeLayout = "01.02.2006 15:04:05"
	defaultScore = 5
)

//Считает штрафы на основе дедлайнов и конфига deadlines.csv
func TransformPenaltyData(raw string, iLab uint, iTimes uint) string {
	dds, err := getDeadlines()
	if err != nil {
		panic(err)
	}

	var data string

	lines := strings.Split(string(raw), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")

		pen := calcPenalty(dds[strings.ToUpper(parts[iLab])], parts[iTimes:])
		for i := 0; i < len(parts) - 2; i++ {
			data += parts[i] + ";"
		}
		data += pen + "\n"
	}

	return data
}

func stringToMoscowTime(oldTime string) (time.Time, error) {
	newTime, err := time.Parse(time.RFC3339, oldTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("ERROR: failed to parse time: %w", err)
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		panic("location invalid")
	}
	newTime = newTime.In(loc)
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
				break
			}
			temp = append(temp, t)
		}
		dds[strings.ToUpper(line[0])] = temp
	}
	return dds, nil
}

func calcPenalty(dds []time.Time, atsStr []string) string {

	ats := []time.Time{}
	for _, at := range atsStr {
		t, err := stringToMoscowTime(at)
		if err != nil {
			return "exist"
		}
		ats = append(ats, t)
	}

	penalty := +2

	for _, dd := range dds {
		for _, at := range ats {
			if at.After(dd) {
				penalty--
			}
		}
	}

	penalty += defaultScore

	return strconv.Itoa(penalty)
}