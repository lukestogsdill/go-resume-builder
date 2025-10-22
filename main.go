package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"gopkg.in/yaml.v3"
)

type Resume struct {
	PersonalInfo PersonalInfo `json:"personal_info" yaml:"personal_info"`
	Summary      string       `json:"summary" yaml:"summary"`
	Experience   []Experience `json:"experience" yaml:"experience"`
	Education    []Education  `json:"education" yaml:"education"`
	Skills       []string     `json:"skills" yaml:"skills"`
}

type PersonalInfo struct {
	Name    string `json:"name" yaml:"name"`
	Email   string `json:"email" yaml:"email"`
	Phone   string `json:"phone" yaml:"phone"`
	Address string `json:"address" yaml:"address"`
	Website string `json:"website" yaml:"website"`
}

type Experience struct {
	Title       string   `json:"title" yaml:"title"`
	Company     string   `json:"company" yaml:"company"`
	Location    string   `json:"location" yaml:"location"`
	StartDate   string   `json:"start_date" yaml:"start_date"`
	EndDate     string   `json:"end_date" yaml:"end_date"`
	Description []string `json:"description" yaml:"description"`
}

type Education struct {
	Degree      string `json:"degree" yaml:"degree"`
	Institution string `json:"institution" yaml:"institution"`
	Location    string `json:"location" yaml:"location"`
	StartDate   string `json:"start_date" yaml:"start_date"`
	EndDate     string `json:"end_date" yaml:"end_date"`
}

func main() {
	var inputFile = flag.String("input", "", "Input resume file (JSON or YAML)")
	var outputFile = flag.String("output", "resume.pdf", "Output PDF file")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: resume-builder -input <file> [-output <file>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	resume, err := loadResume(*inputFile)
	if err != nil {
		log.Fatalf("Error loading resume: %v", err)
	}

	err = generatePDF(resume, *outputFile)
	if err != nil {
		log.Fatalf("Error generating PDF: %v", err)
	}

	fmt.Printf("Resume PDF generated: %s\n", *outputFile)
}

func loadResume(filename string) (*Resume, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var resume Resume
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &resume)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &resume)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return &resume, err
}

func generatePDF(resume *Resume, filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Header - Name
	pdf.SetFont("Arial", "B", 24)
	pdf.Cell(0, 15, resume.PersonalInfo.Name)
	pdf.Ln(20)

	// Contact Info
	pdf.SetFont("Arial", "", 11)
	if resume.PersonalInfo.Email != "" {
		pdf.Cell(0, 6, "Email: "+resume.PersonalInfo.Email)
		pdf.Ln(6)
	}
	if resume.PersonalInfo.Phone != "" {
		pdf.Cell(0, 6, "Phone: "+resume.PersonalInfo.Phone)
		pdf.Ln(6)
	}
	if resume.PersonalInfo.Address != "" {
		pdf.Cell(0, 6, "Address: "+resume.PersonalInfo.Address)
		pdf.Ln(6)
	}
	if resume.PersonalInfo.Website != "" {
		pdf.Cell(0, 6, "Website: "+resume.PersonalInfo.Website)
		pdf.Ln(6)
	}
	pdf.Ln(10)

	// Summary
	if resume.Summary != "" {
		addSection(pdf, "SUMMARY")
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(0, 6, resume.Summary, "", "", false)
		pdf.Ln(10)
	}

	// Experience
	if len(resume.Experience) > 0 {
		addSection(pdf, "EXPERIENCE")
		for _, exp := range resume.Experience {
			// Job title and company
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, exp.Title+" - "+exp.Company)
			pdf.Ln(8)
			
			// Location and dates
			pdf.SetFont("Arial", "I", 10)
			locationAndDates := exp.Location
			if exp.StartDate != "" || exp.EndDate != "" {
				locationAndDates += " | " + exp.StartDate + " - " + exp.EndDate
			}
			pdf.Cell(0, 6, locationAndDates)
			pdf.Ln(8)
			
			// Description
			pdf.SetFont("Arial", "", 10)
			for _, desc := range exp.Description {
				pdf.Cell(5, 5, "•")
				pdf.MultiCell(0, 5, desc, "", "", false)
				pdf.Ln(2)
			}
			pdf.Ln(5)
		}
	}

	// Education
	if len(resume.Education) > 0 {
		addSection(pdf, "EDUCATION")
		for _, edu := range resume.Education {
			pdf.SetFont("Arial", "B", 12)
			pdf.Cell(0, 8, edu.Degree+" - "+edu.Institution)
			pdf.Ln(8)
			
			pdf.SetFont("Arial", "I", 10)
			locationAndDates := edu.Location
			if edu.StartDate != "" || edu.EndDate != "" {
				locationAndDates += " | " + edu.StartDate + " - " + edu.EndDate
			}
			pdf.Cell(0, 6, locationAndDates)
			pdf.Ln(10)
		}
	}

	// Skills
	if len(resume.Skills) > 0 {
		addSection(pdf, "SKILLS")
		pdf.SetFont("Arial", "", 11)
		skillsText := strings.Join(resume.Skills, " • ")
		pdf.MultiCell(0, 6, skillsText, "", "", false)
	}

	return pdf.OutputFileAndClose(filename)
}

func addSection(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, title)
	pdf.Ln(10)
	
	// Add a line under the section title
	pdf.SetLineWidth(0.5)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(8)
}