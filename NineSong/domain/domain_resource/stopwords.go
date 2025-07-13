package domain_resource

import "embed"

//go:embed stopwords/*.txt
var StopWordsFS embed.FS
