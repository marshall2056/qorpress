package main

import (
	"fmt"

	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/storage/gorm/generator"
)

func main() {
	fmt.Println("ready to generate")
	modelpath := "./pkg/models"
	generator.Generate(modelpath, config.Config.Schema)
}
