package pkg

type PaneConfig struct {
	Separator               string                   `json:"separator"`
	SeparatorColor          string                   `json:"separator_color"`
	BackgroundColor         string                   `json:"background_color"`
	OperatorBackgroundColor string                   `json:"operator_background_color"`
	OperationsLeft          []map[string]interface{} `json:"operations_left"`
	OperationsRight         []map[string]interface{} `json:"operations_right"`
}

type Config struct {
	PaneConfig PaneConfig `json:"pane_config"`
	// PromptOperations []Operation
}
