package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/config"

	"resume-builder/utils"
	"resume-builder/templates"
)

func main() {
	var outputFile = flag.String("output", "resume.pdf", "Output PDF file")
	var templateName = flag.String("template", "template-1", "Template to use")
	flag.Parse()

	// Load config and content
	config, err := utils.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	content, err := utils.LoadContent("cnt.json")
	if err != nil {
		log.Fatalf("Error loading content: %v", err)
	}

	// Generate PDF using template
	err = generatePDF(config, content, *templateName, *outputFile)
	if err != nil {
		log.Fatalf("Error generating PDF: %v", err)
	}

	fmt.Printf("Resume PDF generated: %s\n", *outputFile)
}

func generatePDF(cfg *utils.Config, content *utils.Content, templateName, filename string) error {
	// Create maroto config
	cfgBuilder := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(cfg.PDF.Margins.Left).
		WithTopMargin(cfg.PDF.Margins.Top).
		WithRightMargin(cfg.PDF.Margins.Right).
		WithBottomMargin(cfg.PDF.Margins.Bottom)
	
	marotoCfg := cfgBuilder.Build()
	mrt := maroto.New(marotoCfg)

	// Use template-specific logic
	switch templateName {
	case "template-1":
		err := templates.BuildTemplate1(mrt, cfg, content)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown template: %s", templateName)
	}

	// Generate and save
	document, err := mrt.Generate()
	if err != nil {
		return err
	}

	return document.Save(filename)
}