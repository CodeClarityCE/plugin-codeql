package types

type CodeQL struct {
	Runs []Run `json:"runs"`
}

type Run struct {
	Results []Result `json:"results"`
}

type Result struct {
	RuleId    string     `json:"ruleId"`
	RuleIndex int        `json:"ruleIndex"`
	Message   Message    `json:"message"`
	Locations []Location `json:"locations"`
}

type Message struct {
	Text string `json:"text"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseId string `json:"uriBaseId"`
	Index     int    `json:"index"`
}

type Region struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine"`
	StartColumn int `json:"startColumn"`
	EndColumn   int `json:"endColumn"`
}
