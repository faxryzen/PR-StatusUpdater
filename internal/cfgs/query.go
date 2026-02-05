package cfgs

//Получить querys - первый для мержед, второй для открытых репозиториев
func GetQuerys(currentRepo []string) ([]string) {
var (
	queryM     = `
	query {
		repository(owner: "` + currentRepo[1] + `", name: "` + currentRepo[0] + `") {
			pullRequests(states: MERGED, first: 100) {
				nodes {
					number
					title
					author { login }
					mergedAt
					createdAt
					timelineItems(itemTypes: LABELED_EVENT, last: 10) {
						nodes {
							... on LabeledEvent {
								createdAt
								label { name }
							}
						}
					}
					labels(first: 10) {
						nodes {	name }
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
					createdAt
					timelineItems(itemTypes: LABELED_EVENT, last: 10) {
						nodes {
							... on LabeledEvent {
								createdAt
								label { name }
							}
						}
					}
					labels(first: 10) {
						nodes {	name }
					}
				}
			}
		}
	}
	`
	)
	return []string{queryM, queryO}
}
