.data.repository.pullRequests.nodes[] | {
	number: .number,
	author: .author.login,
	labID: (.title | split("/") | .[1]),
	times: {
		created: .createdAt,
		fined: (if (last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt) // null) != null 
			then last(.timelineItems.nodes[] | select(.label.name == "fine").createdAt) 
			else null end),
		merged: (if .mergedAt then .mergedAt else null end)
	},
	marks: [.labels.nodes[].name | select(test("^[+-]\\d+$"))]
}