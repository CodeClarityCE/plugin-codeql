package types

type CodeQL struct {
	Runs []Run `json:"runs"`
}

type Run struct {
	Results []Result `json:"results"`
}

type Result struct {
	RuleId    string  `json:"ruleId"`
	RuleIndex int     `json:"ruleIndex"`
	Message   Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
}
