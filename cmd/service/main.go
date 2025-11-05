package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/faxryzen/pr-updater/internal/cfgs"
	"github.com/faxryzen/pr-updater/internal/fmtc"
)

const (
	fileFormat = ".csv"
	timeLayout = "2006-01-02 15:04:05"
)

func convertTimeToMSK(oldTime string) string {
	oldTime = strings.ReplaceAll(oldTime, "T", " ")
	oldTime = strings.ReplaceAll(oldTime, "Z", "")

	newTime, err := time.Parse(timeLayout, oldTime)
	if err != nil {
		return "null"
	}

	return newTime.Format("2006.01.02 15:04:05")
}

func main() {


	tmpDir, err := os.MkdirTemp("", "prupdater-")
	if err != nil {
		log.Fatal(err)
	}
	os.MkdirAll(tmpDir, 0755)

	tmpFile, err := os.CreateTemp(tmpDir, "prlist_*.csv")
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	currentRepo, err := cfgs.GetRepo()
	if err != nil {
		log.Println(err)

		return
	}

	var (
	queryM     = `
	query {
		repository(owner: "` + currentRepo[1] + `", name: "` + currentRepo[0] + `") {
			pullRequests(states: MERGED, first: 100) {
				nodes {
					number
					title
					author { login }
					timelineItems(itemTypes: LABELED_EVENT, last: 10) {
						nodes {
							... on LabeledEvent {
								createdAt
								label { name }
							}
						}
					}
				}
			}
		}
	}
	`
	queryO     = `
	query {
		repository(owner: "` + currentRepo[1] + `", name: "` + currentRepo[0] + `") {
			pullRequests(states: OPEN, first: 100) {
				nodes {
					number
					title
					author { login }
					timelineItems(itemTypes: LABELED_EVENT, last: 10) {
						nodes {
							... on LabeledEvent {
								createdAt
								label { name }
							}
						}
					}
				}
			}
		}
	}
	`
)

	//номер;логин;лаб;лейбл/статус(fine,merged(fine уже есть),open(т.е fine нету но пр есть));когда был выдан fine (или null если не был)

	fmtc.Cyan.Println("Getting list of Pull Requests from github repo \"" + currentRepo[0] + "\"...")

	cmd := exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", queryM), "--jq",
											`.data.repository.pullRequests.nodes[] | [
											.number,
											.author.login,
											(.title | split("/") | .[1]),
											"merged",
											(if any(.timelineItems.nodes[]; .label.name == "fine") 
											then last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt) 
											else "null" end)
											] | @csv`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	prListMerged := strings.ReplaceAll(string(output), "\"", "")
	prListMerged = strings.ReplaceAll(prListMerged, ",", ";")

	cmd = exec.Command("gh", "api", "graphql", "-f", fmt.Sprintf("query=%s", queryO), "--jq",
											`.data.repository.pullRequests.nodes[] | [
											.number,
											.author.login,
											(.title | split("/") | .[1]),
											"open",
											(if any(.timelineItems.nodes[]; .label.name == "fine") 
											then last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt) 
											else "null" end)
											] | @csv`)

	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	prListOpened := strings.ReplaceAll(string(output), "\"", "")
	prListOpened = strings.ReplaceAll(prListOpened, ",", ";")

	allDataRaw := prListMerged + prListOpened
	var allData string

	lines := strings.Split(string(allDataRaw), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ";")
		parts[4] = convertTimeToMSK(parts[4])
		allData += parts[0]
		for i := 1; i < 5; i++ {
			allData += ";" + parts[i]
		}
		allData += "\n"
	}

	fmtc.Cyan.Println("Choose one of the gists to edit (copy ID) OR leave empty if you want to create new")
	fmt.Println()

	output, err = exec.Command("gh", "gist", "list").Output()
	if err != nil {
		log.Println("ERROR: gh list error")

		return
	}
	fmt.Println(string(output))

	var inputGist string
	fmt.Print("> ")
	_, err = fmt.Scanln(&inputGist)
	if err != nil {
		log.Println("ERROR: scan error")

		return
	}

	output, err = exec.Command("gh", "gist", "view", "--files", inputGist).Output()
	if err != nil {
		log.Println("ERROR: no gist with this ID")

		return
	}

	viewPr := string(output)
	lines = strings.Split(viewPr, "\n")

	gistFiles := []string{}

	fmtc.Green.Println("Founded files in gist:")

	for _, line := range lines {
		if strings.Contains(line, fileFormat) {
			gistFiles = append(gistFiles, line)
			fmtc.Cyan.Println(line)
		}
	}

	for i := range len(gistFiles) {
		if strings.Contains(tmpFile.Name(), gistFiles[i]) {
			fmt.Println("ERROR: Same file name in the gist")

			return
		}
	}

	_, err = tmpFile.WriteString(allData)
	if err != nil {
		log.Println("ERROR: writing in temp")

		return
	}

	cmd = exec.Command("gh", "gist", "edit", inputGist, "--add", tmpFile.Name())
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	fmtc.Green.Println("Successfully added " + tmpFile.Name())

	for i := range len(gistFiles) {
		fmtc.Cyan.Printf("Removing %s\n", gistFiles[i])
		cmd = exec.Command("gh", "gist", "edit", inputGist, "--remove", gistFiles[i])
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("ERROR: %v", err)
			log.Printf("OUTPUT: %s", output)
			return
		}
	}

	fmtc.Green.Println("Done!")
}
