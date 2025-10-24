package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	stdimage "image"
	"image/png"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/linestyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Universal Config Structure
type Config struct {
	PDF              PDFSettings                    `json:"pdf"`
	Spacing          map[string]float64             `json:"spacing"`
	Fonts            map[string]FontDefinition      `json:"fonts"`
	Colors           map[string]string              `json:"colors"`
	Icons            IconConfig                     `json:"icons"`
	SectionTemplates map[string]SectionTemplate     `json:"section_templates"`
	Sections         map[string]SectionConfig       `json:"sections"`
}

type PDFSettings struct {
	PageSize        string  `json:"page_size"`
	Margins         Margins `json:"margins"`
	BackgroundColor string  `json:"background_color"`
}

type Margins struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

type FontDefinition struct {
	Family string  `json:"family"`
	Size   float64 `json:"size"`
	Style  string  `json:"style"`
	Color  string  `json:"color"`
}

type IconConfig struct {
	SVGPaths    []string          `json:"svg_paths"`
	OutputDir   string            `json:"output_dir"`
	DefaultSize int               `json:"default_size"`
	Color       string            `json:"color"`
	Mappings    map[string]string `json:"mappings"`
}

type SectionTemplate struct {
	Spacing      string `json:"spacing"`
	Font         string `json:"font"`
	IconSize     int    `json:"icon_size,omitempty"`
	TitleSpacing string `json:"title_spacing,omitempty"`
	ItemSpacing  string `json:"item_spacing,omitempty"`
}

type SectionConfig struct {
	Template string  `json:"template"`
	Title    string  `json:"title,omitempty"`
	Icon     *string `json:"icon"`
	Enabled  bool    `json:"enabled"`
}

// Content Structure
type Content struct {
	Personal      PersonalInfo           `json:"personal"`
	ContactFields []ContactField         `json:"contact_fields"`
	Sections      map[string]interface{} `json:"sections"`
}

type PersonalInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Website  string `json:"website"`
	GitHub   string `json:"github"`
	LinkedIn string `json:"linkedin"`
}

type ContactField struct {
	Field   string  `json:"field"`
	Content string  `json:"content"`
	Icon    string  `json:"icon"`
	Link    *string `json:"link,omitempty"`
	Type    *string `json:"type,omitempty"`
}

type SectionData struct {
	Content *string     `json:"content,omitempty"`
	Items   []EntryItem `json:"items,omitempty"`
}

type EntryItem struct {
	// For experience/education
	Title       *string  `json:"title,omitempty"`
	Company     *string  `json:"company,omitempty"`
	Institution *string  `json:"institution,omitempty"`
	Degree      *string  `json:"degree,omitempty"`
	Location    *string  `json:"location,omitempty"`
	StartDate   *string  `json:"start_date,omitempty"`
	EndDate     *string  `json:"end_date,omitempty"`
	Description []string `json:"description,omitempty"`
	
	// For simple lists (skills as strings)
	Skill *string `json:"skill,omitempty"`
}

// Universal helper functions
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func LoadContent(filename string) (*Content, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var content Content
	err = json.Unmarshal(data, &content)
	return &content, err
}

func HexToColor(hex string) props.Color {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return props.Color{Red: 0, Green: 0, Blue: 0}
	}
	
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

func ResolveColor(colorName string, colors map[string]string) props.Color {
	if colorValue, exists := colors[colorName]; exists {
		return HexToColor(colorValue)
	}
	return HexToColor(colorName)
}

func EnsureIconExists(iconName string, iconConfig *IconConfig, colors map[string]string) string {
	// Always use colored icons with the configured color
	colorHex := ""
	if iconConfig.Color != "" {
		if colorValue, exists := colors[iconConfig.Color]; exists {
			colorHex = colorValue
		} else {
			colorHex = iconConfig.Color
		}
	}
	
	return EnsureColoredIconExists(iconName, iconConfig, colorHex)
}

func EnsureColoredIconExists(iconName string, iconConfig *IconConfig, colorHex string) string {
	if iconName == "" {
		return ""
	}
	
	iconFileName, exists := iconConfig.Mappings[iconName]
	if !exists {
		return ""
	}
	
	// Always include color in filename
	colorSuffix := strings.TrimPrefix(colorHex, "#")
	if colorSuffix == "" {
		colorSuffix = "000000" // default to black if no color specified
	}
	pngPath := filepath.Join(iconConfig.OutputDir, fmt.Sprintf("%s_%s_%dpx.png", iconFileName, colorSuffix, iconConfig.DefaultSize))
	
	// Return if PNG already exists
	if _, err := os.Stat(pngPath); err == nil {
		return pngPath
	}
	
	// Convert from SVG with optional color
	err := convertColoredIcon(iconFileName, iconConfig.SVGPaths, iconConfig.OutputDir, iconConfig.DefaultSize, colorHex)
	if err != nil {
		log.Printf("Could not convert icon %s: %v", iconName, err)
		return ""
	}
	
	return pngPath
}

func convertIcon(iconName string, svgPaths []string, outputDir string, size int) error {
	return convertColoredIcon(iconName, svgPaths, outputDir, size, "")
}

func convertColoredIcon(iconName string, svgPaths []string, outputDir string, size int, colorHex string) error {
	os.MkdirAll(outputDir, 0755)
	
	for _, svgPath := range svgPaths {
		fullSvgPath := filepath.Join(svgPath, iconName+".svg")
		if _, err := os.Stat(fullSvgPath); err == nil {
			colorSuffix := strings.TrimPrefix(colorHex, "#")
			if colorSuffix == "" {
				colorSuffix = "000000"
			}
			outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s_%dpx.png", iconName, colorSuffix, size))
			return convertSVGToPNG(fullSvgPath, outputPath, size, colorHex)
		}
	}
	
	return fmt.Errorf("icon %s not found in any configured SVG paths", iconName)
}

func convertSVGToPNG(svgPath, pngPath string, maxSize int, colorHex ...string) error {
	svgData, err := os.ReadFile(svgPath)
	if err != nil {
		return fmt.Errorf("failed to read SVG file: %w", err)
	}

	svgContent := string(svgData)
	
	// Apply color if specified
	if len(colorHex) > 0 && colorHex[0] != "" {
		color := strings.TrimPrefix(colorHex[0], "#")
		if len(color) == 6 {
			// Replace existing fill attributes with the new color
			svgContent = strings.ReplaceAll(svgContent, `fill="currentColor"`, `fill="#`+color+`"`)
			svgContent = strings.ReplaceAll(svgContent, `fill='currentColor'`, `fill="#`+color+`"`)
			svgContent = strings.ReplaceAll(svgContent, `fill="#000"`, `fill="#`+color+`"`)
			svgContent = strings.ReplaceAll(svgContent, `fill="#000000"`, `fill="#`+color+`"`)
			svgContent = strings.ReplaceAll(svgContent, `fill="black"`, `fill="#`+color+`"`)
			
			// For FontAwesome icons that have no fill attribute, add one to path elements
			if !strings.Contains(strings.ToLower(svgContent), `fill="`) && !strings.Contains(strings.ToLower(svgContent), `fill='`) {
				// Case insensitive replacement for different SVG formats
				svgContent = strings.ReplaceAll(svgContent, "<path ", `<path fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<Path ", `<Path fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<PATH ", `<PATH fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<circle ", `<circle fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<Circle ", `<Circle fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<rect ", `<rect fill="#`+color+`" `)
				svgContent = strings.ReplaceAll(svgContent, "<polygon ", `<polygon fill="#`+color+`" `)
			}
		}
	}

	icon, err := oksvg.ReadIconStream(strings.NewReader(svgContent))
	if err != nil {
		return fmt.Errorf("failed to parse SVG: %w", err)
	}

	viewBox := icon.ViewBox
	svgWidth := viewBox.W
	svgHeight := viewBox.H
	
	var width, height int
	if svgWidth >= svgHeight {
		width = maxSize
		height = int(float64(maxSize) * svgHeight / svgWidth)
	} else {
		height = maxSize
		width = int(float64(maxSize) * svgWidth / svgHeight)
	}
	
	w, h := float64(width), float64(height)
	icon.SetTarget(0, 0, w, h)

	img := stdimage.NewRGBA(stdimage.Rect(0, 0, width, height))
	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)
	
	icon.Draw(raster, 1.0)

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

func ResolveTemplate(templateStr string, content *Content) string {
	if !strings.Contains(templateStr, "{{") {
		return templateStr
	}
	
	tmpl, err := template.New("content").Parse(templateStr)
	if err != nil {
		return templateStr
	}
	
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, content)
	if err != nil {
		return templateStr
	}
	
	return buf.String()
}

func ResolveFontStyle(styleStr string) fontstyle.Type {
	switch strings.ToLower(styleStr) {
	case "bold":
		return fontstyle.Bold
	case "italic":
		return fontstyle.Italic
	case "bolditalic":
		return fontstyle.BoldItalic
	default:
		return fontstyle.Normal
	}
}

func AddDivider(mrt core.Maroto, cfg *Config) error {
	dividerColor := ResolveColor("secondary", cfg.Colors)
	
	lineComp := line.New(props.Line{
		Color:         &dividerColor,
		Style:         linestyle.Solid,
		Thickness:     1.0,
		Orientation:   orientation.Horizontal,
		OffsetPercent: 50.0,
		SizePercent:   100.0,
	})

	mrt.AddRow(cfg.Spacing["medium"], col.New(12).Add(lineComp))
	return nil
}

func CreateColoredIcon(iconName string, cfg *Config, size int) core.Component {
	iconPath := EnsureIconExists(iconName, &cfg.Icons, cfg.Colors)
	if iconPath == "" {
		log.Printf("Could not create icon for %s", iconName)
		// Return empty component if icon can't be created
		return text.New("", props.Text{})
	}
	
	return image.NewFromFile(iconPath, props.Rect{
		Left:    0.0,
		Top:     1.0,
		Percent: float64(size),
	})
}

func AddSectionTitle(mrt core.Maroto, cfg *Config, sectionCfg SectionConfig, template SectionTemplate) error {
	titleFont := cfg.Fonts["section_title"]
	titleColor := ResolveColor(titleFont.Color, cfg.Colors)
	titleSpacing := cfg.Spacing[template.TitleSpacing]

	var cols []core.Col
	
	// Add icon if specified
	if sectionCfg.Icon != nil && *sectionCfg.Icon != "" {
		iconSize := template.IconSize
		if iconSize == 0 {
			iconSize = 60 // default fallback
		}
		iconImg := CreateColoredIcon(*sectionCfg.Icon, cfg, iconSize)
		cols = append(cols, col.New(1).Add(iconImg))
	}

	// Add title text
	titleCol := 11
	if len(cols) > 0 {
		titleCol = 11
	} else {
		titleCol = 12
	}

	titleComp := text.New(sectionCfg.Title, props.Text{
		Family: titleFont.Family,
		Size:   titleFont.Size,
		Style:  ResolveFontStyle(titleFont.Style),
		Color:  &titleColor,
		Align:  align.Left,
		Left:   3.0,
	})
	
	cols = append(cols, col.New(titleCol).Add(titleComp))

	mrt.AddRow(titleSpacing, cols...)
	return nil
}

func CreateContactFieldCol(field ContactField, cfg *Config, content *Content) core.Col {
	contactFont := cfg.Fonts[cfg.SectionTemplates["contact"].Font]
	
	// Resolve template variables
	fieldContent := ResolveTemplate(field.Content, content)
	if fieldContent == "" {
		return col.New(4)
	}

	iconPath := EnsureIconExists(field.Icon, &cfg.Icons, cfg.Colors)
	
	var mainComp core.Component
	
	// Create hyperlink if specified
	if field.Type != nil && *field.Type == "link" && field.Link != nil {
		linkURL := ResolveTemplate(*field.Link, content)
		linkColor := ResolveColor("link", cfg.Colors)
		mainComp = text.New(fieldContent, props.Text{
			Family:    contactFont.Family,
			Size:      contactFont.Size,
			Top:       1.0,
			Left:      7.0,
			Color:     &linkColor,
			Style:     ResolveFontStyle(contactFont.Style),
			Hyperlink: &linkURL,
		})
	} else {
		textColor := ResolveColor(contactFont.Color, cfg.Colors)
		mainComp = text.New(fieldContent, props.Text{
			Family: contactFont.Family,
			Size:   contactFont.Size,
			Top:    1.0,
			Left:   7.0,
			Color:  &textColor,
			Style:  ResolveFontStyle(contactFont.Style),
		})
	}

	if iconPath != "" {
		contactTemplate := cfg.SectionTemplates["contact"]
		iconSize := contactTemplate.IconSize
		if iconSize == 0 {
			iconSize = 60 // default fallback
		}
		iconImg := CreateColoredIcon(field.Icon, cfg, iconSize)
		return col.New(4).Add(iconImg, mainComp)
	} else {
		return col.New(4).Add(mainComp)
	}
}