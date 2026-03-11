.data.repository.pullRequests.nodes[] | {
	number: .number,
	author: .author.login,
	name: (.title | capture("\\[(?<content>[^\\]]*)\\]").content // ""),
	times: {
		created: .createdAt,
		merged: (if .mergedAt then .mergedAt else null end)
	},
}