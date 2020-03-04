package generator

// Model : Struct for each model
type Model struct {
	Name       string   `yaml:"name"`
	ColumnList []Column `yaml:"columnList"`
}

// Column : Struct to define each column in model
type Column struct {
	Name     string    `yaml:"name"`
	DataType string    `yaml:"dataType"`
	GormTag  []TagName `yaml:"gormTag"`
	JSONTag  string    `yaml:"jsonTag"`
	FormTag  string    `yaml:"formTag"`
}

// TagName : These are tag name for grom tags used in column struct
type TagName struct {
	Tag string `yaml:"tag"`
}
