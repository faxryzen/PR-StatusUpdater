package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fatih/color"
	cfginteraction "github.com/faxryzen/pr-updater/internal/cfg_interaction"
)

const (
	fileFormat = ".csv"
	timeLayout = "2006-01-02 15:04:05"
)

var (
	queryM     = `
	query {
		repository(owner: "Volgarenok", name: "spbspu-anp-2025-5130904-50004") {
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
		repository(owner: "Volgarenok", name: "spbspu-anp-2025-5130904-50004") {
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
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)

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

	repos, err := cfginteraction.CfgReadRepoCSV()
	if err != nil {
		log.Println(err)

		return
	}

	cyan.Println("Recent repositories:")
	var repoCounter uint = 0

	yellow.Println("0: Add new repo")
	for i, str := range repos {
		yellow.Printf("%d: %s by %s\n", i + 1, str[0], str[1])
		repoCounter++;
	}

	var inputRepo uint
	fmt.Print("> ")
	_, err = fmt.Scanln(&inputRepo)
	if err != nil {
		log.Println("ERROR: scan error")

		return
	}

	if inputRepo == 0 {
		cyan.Println("Type repo name and owner (use space for separator)")
		fmt.Print("> ")
		var (
			name  string
			owner string
		)
		_, err = fmt.Scanln(&name, &owner)
		if err != nil {
			log.Println("ERROR: scan error")

			return
		}
		inputNewRepo := []string{name, owner}
		err = cfginteraction.CfgWriteRepoCSV(inputNewRepo)
		if err != nil {
			log.Println("ERROR: cfg goes wrong")

			return
		}
		repoCounter++;
		inputRepo = repoCounter
	}

	if inputRepo > repoCounter {
		log.Println("ERROR: wrong number")

		return
	}

	currentRepo := []string{repos[inputRepo - 1][0], repos[inputRepo - 1][1]}

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
		repository(owner: "Volgarenok", name: "spbspu-anp-2025-5130904-50004") {
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

	cyan.Println("Getting list of Pull Requests from github repo...")

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

	//fmt.Println(allData)

	cyan.Println("Choose one of the gists to edit (copy ID) OR leave empty if you want to create new")
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

	green.Println("Founded files in gist:")

	for _, line := range lines {
		if strings.Contains(line, fileFormat) {
			gistFiles = append(gistFiles, line)
			cyan.Println(line)
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

	green.Println("Successfully added " + tmpFile.Name())

	for i := range len(gistFiles) {
		cyan.Printf("Removing %s\n", gistFiles[i])
		cmd = exec.Command("gh", "gist", "edit", inputGist, "--remove", gistFiles[i])
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("ERROR: %v", err)
			log.Printf("OUTPUT: %s", output)
			return
		}
	}

	green.Println("Done!")
}
