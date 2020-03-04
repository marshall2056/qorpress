package models

//go:generate gp-extender -structs Event -output event-funcs.go
type Event struct {
	ID         uint       `gorm:"primary_key"`
	Categories []Category `gorm:"many2many:category_event"`
	Title      string
	Slug       string `gorm:"unique"`
	Body       string `gorm:"type:longtext"`
	Summary    string `gorm:"type:longtext"`
	Images     []Image
	Documents  []Document
	Videos     []Video
	Links      []Link
	Type       string
	Created    int32
	Updated    int32
	StartDate  int32
	EndDate    int32
}
