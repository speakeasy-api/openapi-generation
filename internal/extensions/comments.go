package extensions

type Comment struct {
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description" json:"description"`
}

type Comments map[string]*Comment
