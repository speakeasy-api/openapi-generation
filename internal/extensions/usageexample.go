package extensions

type UsageExampleConfig struct {
	Title       string   `json:"title" yaml:"title,omitempty"`
	Description string   `json:"description" yaml:"description,omitempty"`
	Position    int      `json:"position" yaml:"position,omitempty"`
	Tags        []string `json:"tags" yaml:"tags,omitempty"`
}
