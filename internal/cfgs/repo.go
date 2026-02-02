package cfgs

type Repo struct {
	Name  string
	Owner string
}

// Получить список репозиториев
func GetRepos() ([]Repo, error) {
	records, err := ScanRecords("repos.csv")
	if err != nil {
		return nil, err
	}

	repos := make([]Repo, 0, len(records))
	for _, r := range records {
		repos = append(repos, Repo{
			Name:  r[0],
			Owner: r[1],
		})
	}

	return repos, nil
}

// Сохранить репозиторий в начало списка
func SaveRepo(repo Repo) error {
	records, err := ScanRecords("repos.csv")
	if err != nil {
		return err
	}

	newRecord := []string{repo.Name, repo.Owner}
	records = append([][]string{newRecord}, records...)

	return WriteRecords(records, "repos.csv")
}
