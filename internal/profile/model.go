package profile

type Profile struct {
	Name        string              `json:"name"`
	DisplayName string              `json:"display_name"`
	DataSources []DataSourceSummary `json:"data_sources,omitempty"`
	Tier1       PersonaCore         `json:"tier1"`
	Tier2       ReviewPatterns      `json:"tier2"`
	Tier3       CodebaseRules       `json:"tier3"`
	Tier4       CommunicationVoice  `json:"tier4"`
}

type DataSourceSummary struct {
	Source string `json:"source"`
	Kind   string `json:"kind"`
	Count  int    `json:"count"`
}

type PersonaCore struct {
	Philosophy       string   `json:"philosophy"`
	Priorities       []string `json:"priorities"`
	ApprovalCriteria string   `json:"approval_criteria"`
	PetPeeves        []string `json:"pet_peeves"`
}

type ReviewPatterns struct {
	Categories []ReviewCategory `json:"categories"`
}

type ReviewCategory struct {
	Name         string   `json:"name"`
	TriggerGlobs []string `json:"trigger_globs"`
	Patterns     []string `json:"patterns"`
	Examples     []string `json:"examples"`
}

type CodebaseRules struct {
	Wrappers    []CodeRule `json:"wrappers"`
	BaseClasses []CodeRule `json:"base_classes"`
	Naming      []CodeRule `json:"naming"`
	Patterns    []CodeRule `json:"patterns"`
}

type CodeRule struct {
	Rule    string `json:"rule"`
	Example string `json:"example,omitempty"`
}

type CommunicationVoice struct {
	Tone             string   `json:"tone"`
	PraisePhrases    []string `json:"praise_phrases"`
	FeedbackPhrases  []string `json:"feedback_phrases"`
	RecurringPhrases []string `json:"recurring_phrases"`
}
