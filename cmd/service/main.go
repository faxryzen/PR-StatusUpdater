package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"os/exec"
	"github.com/fatih/color"
)

const (
	repo       = "git@github.com:Volgarenok/spbspu-anp-2025-5130904-50401.git"
	fileFormat = ".csv"
	stateM     = "merged"
	stateO     = "open"
)

func main() {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)

	tmpDir := "/tmp/prupdater"
	os.MkdirAll(tmpDir, 0755)

	tmpFile, err := os.CreateTemp(tmpDir, "prlist_*.csv")
	if err != nil {
		log.Println(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cmd := exec.Command("gh", "pr", "--repo", repo, "ls", "--state", stateM, "--json", "author", "--json", "title", "--jq", ".[] | [\"merged\", .author.login, (.title | split(\"/\") | .[1])] | @csv")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	prListMerged := strings.ReplaceAll(string(output), "\"", "")
	prListMerged = strings.ReplaceAll(prListMerged, ",", ";")

	cmd = exec.Command("gh", "pr", "--repo", repo, "ls", "--state", stateO, "--json", "author", "--json", "title", "--jq", ".[] | [\"open\", .author.login, (.title | split(\"/\") | .[1])] | @csv")

	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	prListOpened := strings.ReplaceAll(string(output), "\"", "")
	prListOpened = strings.ReplaceAll(prListOpened, ",", ";")


	allData := prListMerged + prListOpened

	cyan.Println("Choose one of the gists to edit (copy ID) OR leave empty if you want to create new")
	fmt.Println()

	output, err = exec.Command("gh", "gist", "list").Output()
	if err != nil {
		log.Println("ERROR: gh list error")

		return
	}
	fmt.Println(string(output))

	var input string

	_, err = fmt.Scanln(&input)
	if err != nil {
		log.Println("ERROR: scan error")

		return
	}

	output, err = exec.Command("gh", "gist", "view", "--files", input).Output()
	if err != nil {
		log.Println("ERROR: no gist with this ID")

		return
	}

	viewPr := string(output)
	lines := strings.Split(viewPr, "\n")

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

	cmd = exec.Command("gh", "gist", "edit", input, "--add", tmpFile.Name())
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR: %v", err)
		log.Printf("OUTPUT: %s", output)
		return
	}

	green.Println("Successfully added " + tmpFile.Name())

	for i := range len(gistFiles) {
		cyan.Printf("Removing %s\n", gistFiles[i])
		cmd = exec.Command("gh", "gist", "edit", input, "--remove", gistFiles[i])
		output, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("ERROR: %v", err)
			log.Printf("OUTPUT: %s", output)
			return
		}
	}

	green.Println("Done!")
}
