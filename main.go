package main

import (
	"encoding/json"
	"flag"
	"fmt"
	stdimage "image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"gopkg.in/yaml.v3"
)

type Resume struct {
	PersonalInfo PersonalInfo           `json:"personal_info" yaml:"personal_info"`
	Sections     []Section              `json:"sections" yaml:"sections"`
	Layout       map[string]interface{} `json:"layout" yaml:"layout"`
}

type Section struct {
	Type    string                 `json:"type" yaml:"type"`
	Title   string                 `json:"title" yaml:"title"`
	Order   int                    `json:"order" yaml:"order"`
	Enabled bool                   `json:"enabled" yaml:"enabled"`
	Data    map[string]interface{} `json:"data" yaml:"data"`
}

type SectionRenderer interface {
	Render(mrt core.Maroto, section Section, style *StyleConfig)
}

type SectionRegistry map[string]SectionRenderer

// Section Types
type TextSection struct{}
type ExperienceSection struct{}
type EducationSection struct{}
type SkillsSection struct{}
type ContactSection struct{}

type StyleConfig struct {
	Colors  ColorScheme `json:"colors" yaml:"colors"`
	Fonts   FontConfig  `json:"fonts" yaml:"fonts"`
	Spacing SpacingConfig `json:"spacing" yaml:"spacing"`
	Layout  LayoutConfig `json:"layout" yaml:"layout"`
}

type ColorScheme struct {
	Primary    string `json:"primary" yaml:"primary"`
	Secondary  string `json:"secondary" yaml:"secondary"`
	Text       string `json:"text" yaml:"text"`
	Background string `json:"background" yaml:"background"`
	Accent     string `json:"accent" yaml:"accent"`
}

type FontConfig struct {
	Header   []string `json:"header" yaml:"header"`
	Body     []string `json:"body" yaml:"body"`
	Fallback string   `json:"fallback" yaml:"fallback"`
}

type SpacingConfig struct {
	SectionGap    float64 `json:"section_gap" yaml:"section_gap"`
	ItemGap       float64 `json:"item_gap" yaml:"item_gap"`
	LineHeight    float64 `json:"line_height" yaml:"line_height"`
	BulletIndent  float64 `json:"bullet_indent" yaml:"bullet_indent"`
}

type LayoutConfig struct {
	Margins      Margins `json:"margins" yaml:"margins"`
	PageSize     string  `json:"page_size" yaml:"page_size"`
	HeaderSize   float64 `json:"header_size" yaml:"header_size"`
	BodySize     float64 `json:"body_size" yaml:"body_size"`
	SinglePage   bool    `json:"single_page" yaml:"single_page"`
	CompactMode  bool    `json:"compact_mode" yaml:"compact_mode"`
}

type Margins struct {
	Top    float64 `json:"top" yaml:"top"`
	Bottom float64 `json:"bottom" yaml:"bottom"`
	Left   float64 `json:"left" yaml:"left"`
	Right  float64 `json:"right" yaml:"right"`
}

type PersonalInfo struct {
	Name     string            `json:"name" yaml:"name"`
	Email    string            `json:"email" yaml:"email"`
	Phone    string            `json:"phone" yaml:"phone"`
	Address  string            `json:"address" yaml:"address"`
	Website  string            `json:"website" yaml:"website"`
	GitHub   string            `json:"github" yaml:"github"`
	LinkedIn string            `json:"linkedin" yaml:"linkedin"`
	Icons    map[string]string `json:"icons" yaml:"icons"`
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

// Font Awesome PNG icon paths
var fontAwesomeIcons = map[string]string{
	"email":    "icons/envelope.png",
	"phone":    "icons/phone.png",
	"website":  "icons/globe.png",
	"github":   "icons/github.png",
	"linkedin": "icons/linkedin.png",
	"address":  "icons/location-dot.png",
	"location": "icons/location-dot.png",
}

func main() {
	var inputFile = flag.String("input", "", "Input resume file (JSON or YAML)")
	var outputFile = flag.String("output", "resume.pdf", "Output PDF file")
	var styleFile = flag.String("style", "", "Style configuration file (JSON or YAML)")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Usage: resume-builder -input <file> [-output <file>] [-style <file>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	resume, err := loadResume(*inputFile)
	if err != nil {
		log.Fatalf("Error loading resume: %v", err)
	}

	style := getDefaultStyle()
	if *styleFile != "" {
		loadedStyle, err := loadStyle(*styleFile)
		if err != nil {
			log.Fatalf("Error loading style: %v", err)
		}
		style = loadedStyle
	}

	err = generatePDF(resume, style, *outputFile)
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

func loadStyle(filename string) (*StyleConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var style StyleConfig
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &style)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &style)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return &style, err
}

func getDefaultStyle() *StyleConfig {
	return &StyleConfig{
		Colors: ColorScheme{
			Primary:    "#2C3E50",
			Secondary:  "#34495E", 
			Text:       "#2C3E50",
			Background: "#FFFFFF",
			Accent:     "#3498DB",
		},
		Fonts: FontConfig{
			Header:   []string{"Helvetica", "Arial"},
			Body:     []string{"Helvetica", "Arial"},
			Fallback: "Arial",
		},
		Spacing: SpacingConfig{
			SectionGap:   15.0,
			ItemGap:      8.0,
			LineHeight:   6.0,
			BulletIndent: 5.0,
		},
		Layout: LayoutConfig{
			Margins: Margins{
				Top:    20.0,
				Bottom: 20.0,
				Left:   20.0,
				Right:  20.0,
			},
			PageSize:    "A4",
			HeaderSize:  24.0,
			BodySize:    11.0,
			SinglePage:  false,
			CompactMode: false,
		},
	}
}

func hexToColor(hex string) props.Color {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")
	
	// Default to black if invalid
	if len(hex) != 6 {
		return props.Color{Red: 0, Green: 0, Blue: 0}
	}
	
	// Parse hex values
	r := parseInt(hex[0:2])
	g := parseInt(hex[2:4])
	b := parseInt(hex[4:6])
	
	return props.Color{Red: r, Green: g, Blue: b}
}

func parseInt(hex string) int {
	result := 0
	for _, char := range hex {
		result *= 16
		if char >= '0' && char <= '9' {
			result += int(char - '0')
		} else if char >= 'a' && char <= 'f' {
			result += int(char - 'a' + 10)
		} else if char >= 'A' && char <= 'F' {
			result += int(char - 'A' + 10)
		}
	}
	return result
}

func addContactWithSVG(mrt core.Maroto, iconKey string, contactText string, iconName string, textColor props.Color, fontSize float64, rowHeight float64) {
	if contactText == "" {
		return
	}
	
	// Use custom icon name if provided, otherwise use default mapping
	var iconPath string
	if iconName != "" {
		iconPath = ensureIconExists(iconName)
	} else if defaultPath, exists := fontAwesomeIcons[iconKey]; exists {
		iconPath = defaultPath
	}
	
	if iconPath == "" {
		// Fallback to text only
		mrt.AddRow(rowHeight,
			col.New(12).Add(
				text.New(contactText, props.Text{
					Size:  fontSize,
					Color: &textColor,
				}),
			),
		)
		return
	}
	
	// Add row with icon + text using maroto pattern
	mrt.AddRow(rowHeight,
		image.NewFromFileCol(1, iconPath),
		col.New(11).Add(
			text.New(contactText, props.Text{
				Size:  fontSize,
				Color: &textColor,
				Top:   1,
			}),
		),
	)
}

// ensureIconExists converts SVG to PNG if needed and returns PNG path
func ensureIconExists(iconName string) string {
	pngPath := filepath.Join("icons", iconName+".png")
	
	// Return if PNG already exists
	if _, err := os.Stat(pngPath); err == nil {
		return pngPath
	}
	
	// Try to convert from Font Awesome SVGs
	err := convertFontAwesomeIcon(iconName)
	if err != nil {
		log.Printf("Could not convert icon %s: %v", iconName, err)
		return ""
	}
	
	return pngPath
}

// convertFontAwesomeIcon finds and converts a Font Awesome icon by name
func convertFontAwesomeIcon(iconName string) error {
	// Ensure icons directory exists
	os.MkdirAll("icons", 0755)
	
	// Try solid icons first
	svgPath := filepath.Join("fontawesome-free-6.4.0-desktop", "svgs", "solid", iconName+".svg")
	if _, err := os.Stat(svgPath); err == nil {
		return convertSVGToPNG(svgPath, filepath.Join("icons", iconName+".png"), 32)
	}
	
	// Try brands icons
	svgPath = filepath.Join("fontawesome-free-6.4.0-desktop", "svgs", "brands", iconName+".svg")
	if _, err := os.Stat(svgPath); err == nil {
		return convertSVGToPNG(svgPath, filepath.Join("icons", iconName+".png"), 32)
	}
	
	return fmt.Errorf("icon %s not found in solid or brands", iconName)
}

func convertSVGToPNG(svgPath, pngPath string, maxSize int) error {
	// Read SVG file
	svgData, err := os.ReadFile(svgPath)
	if err != nil {
		return fmt.Errorf("failed to read SVG file: %w", err)
	}

	// Parse SVG
	icon, err := oksvg.ReadIconStream(strings.NewReader(string(svgData)))
	if err != nil {
		return fmt.Errorf("failed to parse SVG: %w", err)
	}

	// Get SVG dimensions and calculate aspect ratio
	viewBox := icon.ViewBox
	svgWidth := viewBox.W
	svgHeight := viewBox.H
	
	// Calculate dimensions maintaining aspect ratio
	var width, height int
	if svgWidth >= svgHeight {
		width = maxSize
		height = int(float64(maxSize) * svgHeight / svgWidth)
	} else {
		height = maxSize
		width = int(float64(maxSize) * svgWidth / svgHeight)
	}
	
	// Create raster context
	w, h := float64(width), float64(height)
	icon.SetTarget(0, 0, w, h)

	// Create image
	img := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))
	
	// Create scanner and rasterize
	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)
	
	// Draw SVG to raster
	icon.Draw(raster, 1.0)

	// Save as PNG
	outFile, err := os.Create(pngPath)
	if err != nil {
		return fmt.Errorf("failed to create PNG file: %w", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// TextSection renderer (for summary, etc.)
func (ts TextSection) Render(mrt core.Maroto, section Section, style *StyleConfig) {
	textColor := hexToColor(style.Colors.Text)
	secondaryColor := hexToColor(style.Colors.Secondary)
	
	addMarotoSection(mrt, section.Title, secondaryColor)
	
	if content, ok := section.Data["content"].(string); ok {
		mrt.AddRow(8,
			col.New(12).Add(
				text.New(content, props.Text{
					Size:  style.Layout.BodySize,
					Color: &textColor,
					Align: align.Left,
				}),
			),
		)
	}
}

// ExperienceSection renderer
func (es ExperienceSection) Render(mrt core.Maroto, section Section, style *StyleConfig) {
	textColor := hexToColor(style.Colors.Text)
	secondaryColor := hexToColor(style.Colors.Secondary)
	
	addMarotoSection(mrt, section.Title, secondaryColor)
	
	if items, ok := section.Data["items"].([]interface{}); ok {
		for _, item := range items {
			if exp, ok := item.(map[string]interface{}); ok {
				// Job title and company
				title := getString(exp, "title")
				company := getString(exp, "company")
				if title != "" && company != "" {
					mrt.AddRow(8,
						col.New(12).Add(
							text.New(title+" - "+company, props.Text{
								Size:  style.Layout.BodySize + 1,
								Style: fontstyle.Bold,
								Color: &textColor,
							}),
						),
					)
				}
				
				// Location and dates
				location := getString(exp, "location")
				startDate := getString(exp, "start_date")
				endDate := getString(exp, "end_date")
				locationAndDates := location
				if startDate != "" || endDate != "" {
					locationAndDates += " | " + startDate + " - " + endDate
				}
				if locationAndDates != "" {
					mrt.AddRow(6,
						col.New(12).Add(
							text.New(locationAndDates, props.Text{
								Size:  style.Layout.BodySize - 1,
								Style: fontstyle.Italic,
								Color: &textColor,
							}),
						),
					)
				}
				
				// Description bullets
				if desc, ok := exp["description"].([]interface{}); ok {
					for _, bullet := range desc {
						if bulletStr, ok := bullet.(string); ok {
							mrt.AddRow(6,
								col.New(12).Add(
									text.New("• "+bulletStr, props.Text{
										Size:  style.Layout.BodySize - 1,
										Color: &textColor,
									}),
								),
							)
						}
					}
				}
			}
		}
	}
}

// EducationSection renderer
func (eds EducationSection) Render(mrt core.Maroto, section Section, style *StyleConfig) {
	textColor := hexToColor(style.Colors.Text)
	secondaryColor := hexToColor(style.Colors.Secondary)
	
	addMarotoSection(mrt, section.Title, secondaryColor)
	
	if items, ok := section.Data["items"].([]interface{}); ok {
		for _, item := range items {
			if edu, ok := item.(map[string]interface{}); ok {
				degree := getString(edu, "degree")
				institution := getString(edu, "institution")
				if degree != "" && institution != "" {
					mrt.AddRow(8,
						col.New(12).Add(
							text.New(degree+" - "+institution, props.Text{
								Size:  style.Layout.BodySize + 1,
								Style: fontstyle.Bold,
								Color: &textColor,
							}),
						),
					)
				}
				
				location := getString(edu, "location")
				startDate := getString(edu, "start_date")
				endDate := getString(edu, "end_date")
				locationAndDates := location
				if startDate != "" || endDate != "" {
					locationAndDates += " | " + startDate + " - " + endDate
				}
				if locationAndDates != "" {
					mrt.AddRow(6,
						col.New(12).Add(
							text.New(locationAndDates, props.Text{
								Size:  style.Layout.BodySize - 1,
								Style: fontstyle.Italic,
								Color: &textColor,
							}),
						),
					)
				}
			}
		}
	}
}

// SkillsSection renderer
func (ss SkillsSection) Render(mrt core.Maroto, section Section, style *StyleConfig) {
	textColor := hexToColor(style.Colors.Text)
	secondaryColor := hexToColor(style.Colors.Secondary)
	
	addMarotoSection(mrt, section.Title, secondaryColor)
	
	if items, ok := section.Data["items"].([]interface{}); ok {
		var skills []string
		for _, item := range items {
			if skill, ok := item.(string); ok {
				skills = append(skills, skill)
			}
		}
		if len(skills) > 0 {
			skillsText := strings.Join(skills, " • ")
			mrt.AddRow(8,
				col.New(12).Add(
					text.New(skillsText, props.Text{
						Size:  style.Layout.BodySize,
						Color: &textColor,
					}),
				),
			)
		}
	}
}

// Helper function to get string from map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// Create section registry
func createSectionRegistry() SectionRegistry {
	return SectionRegistry{
		"text":       TextSection{},
		"summary":    TextSection{},
		"experience": ExperienceSection{},
		"education":  EducationSection{},
		"skills":     SkillsSection{},
	}
}

func generatePDF(resume *Resume, style *StyleConfig, filename string) error {
	// Create maroto config
	cfgBuilder := config.NewBuilder().
		WithPageNumber().
		WithLeftMargin(style.Layout.Margins.Left).
		WithTopMargin(style.Layout.Margins.Top).
		WithRightMargin(style.Layout.Margins.Right).
		WithBottomMargin(style.Layout.Margins.Bottom)

	// Font Awesome will be handled by fallback to Unicode for now
	// maroto v2 has complex font handling, using Unicode symbols instead

	cfg := cfgBuilder.Build()

	// Create maroto instance
	mrt := maroto.New(cfg)

	// Build PDF content
	buildResumePage(mrt, resume, style)

	// Generate and save
	document, err := mrt.Generate()
	if err != nil {
		return err
	}

	return document.Save(filename)
}

func buildResumePage(mrt core.Maroto, resume *Resume, style *StyleConfig) {
	primaryColor := hexToColor(style.Colors.Primary)
	textColor := hexToColor(style.Colors.Text)

	// Header - Name
	mrt.AddRow(20,
		col.New(12).Add(
			text.New(resume.PersonalInfo.Name, props.Text{
				Top:   5,
				Size:  style.Layout.HeaderSize,
				Style: fontstyle.Bold,
				Color: &primaryColor,
			}),
		),
	)

	// Determine spacing based on layout settings
	rowHeight := 6.0
	if style.Layout.CompactMode || style.Layout.SinglePage {
		rowHeight = 4.0
	}

	// Contact Info with auto-converting icons
	addContactWithSVG(mrt, "email", resume.PersonalInfo.Email, resume.PersonalInfo.Icons["email"], textColor, style.Layout.BodySize, rowHeight)
	addContactWithSVG(mrt, "phone", resume.PersonalInfo.Phone, resume.PersonalInfo.Icons["phone"], textColor, style.Layout.BodySize, rowHeight)
	addContactWithSVG(mrt, "address", resume.PersonalInfo.Address, resume.PersonalInfo.Icons["address"], textColor, style.Layout.BodySize, rowHeight)
	addContactWithSVG(mrt, "website", resume.PersonalInfo.Website, resume.PersonalInfo.Icons["website"], textColor, style.Layout.BodySize, rowHeight)
	addContactWithSVG(mrt, "github", resume.PersonalInfo.GitHub, resume.PersonalInfo.Icons["github"], textColor, style.Layout.BodySize, rowHeight)
	addContactWithSVG(mrt, "linkedin", resume.PersonalInfo.LinkedIn, resume.PersonalInfo.Icons["linkedin"], textColor, style.Layout.BodySize, rowHeight)

	// Dynamic sections
	registry := createSectionRegistry()
	
	// Sort sections by order
	sections := make([]Section, len(resume.Sections))
	copy(sections, resume.Sections)
	
	// Simple sort by order
	for i := 0; i < len(sections); i++ {
		for j := i + 1; j < len(sections); j++ {
			if sections[i].Order > sections[j].Order {
				sections[i], sections[j] = sections[j], sections[i]
			}
		}
	}
	
	// Render each enabled section
	for _, section := range sections {
		if section.Enabled {
			if renderer, exists := registry[section.Type]; exists {
				renderer.Render(mrt, section, style)
			}
		}
	}
}

func addMarotoSection(mrt core.Maroto, title string, headerColor props.Color) {
	mrt.AddRow(12,
		col.New(12).Add(
			text.New(title, props.Text{
				Size:  14,
				Style: fontstyle.Bold,
				Color: &headerColor,
			}),
		),
	)
}