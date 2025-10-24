package templates

import (
	"strings"

	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"resume-builder/utils"
)

// BuildTemplate1 creates a template-1 style resume
func BuildTemplate1(mrt core.Maroto, cfg *utils.Config, content *utils.Content) error {
	// Header Section
	if cfg.Sections["header"].Enabled {
		err := buildHeaderSection(mrt, cfg, content)
		if err != nil {
			return err
		}
	}

	// Contact Section
	if cfg.Sections["contact"].Enabled {
		err := buildContactSection(mrt, cfg, content)
		if err != nil {
			return err
		}
		
		// Add divider after contact
		err = utils.AddDivider(mrt, cfg)
		if err != nil {
			return err
		}
	}

	// Dynamic sections in order
	sectionOrder := []string{"summary", "experience", "education", "skills", "certifications"}
	
	for _, sectionKey := range sectionOrder {
		sectionCfg, exists := cfg.Sections[sectionKey]
		if !exists || !sectionCfg.Enabled {
			continue
		}

		sectionData, hasData := content.Sections[sectionKey]
		if !hasData {
			continue
		}

		switch sectionCfg.Template {
		case "simple_list":
			err := buildSimpleListSection(mrt, cfg, sectionCfg, sectionData)
			if err != nil {
				return err
			}
		case "entry_list":
			// Convert to SectionData struct for entry lists
			var data utils.SectionData
			if dataMap, ok := sectionData.(map[string]interface{}); ok {
				if items, exists := dataMap["items"]; exists {
					// Convert interface{} to []EntryItem
					if itemsSlice, ok := items.([]interface{}); ok {
						for _, item := range itemsSlice {
							if itemMap, ok := item.(map[string]interface{}); ok {
								var entry utils.EntryItem
								// Convert each field
								if title, ok := itemMap["title"].(string); ok {
									entry.Title = &title
								}
								if company, ok := itemMap["company"].(string); ok {
									entry.Company = &company
								}
								if institution, ok := itemMap["institution"].(string); ok {
									entry.Institution = &institution
								}
								if degree, ok := itemMap["degree"].(string); ok {
									entry.Degree = &degree
								}
								if location, ok := itemMap["location"].(string); ok {
									entry.Location = &location
								}
								if startDate, ok := itemMap["start_date"].(string); ok {
									entry.StartDate = &startDate
								}
								if endDate, ok := itemMap["end_date"].(string); ok {
									entry.EndDate = &endDate
								}
								if desc, ok := itemMap["description"].([]interface{}); ok {
									for _, d := range desc {
										if descStr, ok := d.(string); ok {
											entry.Description = append(entry.Description, descStr)
										}
									}
								}
								data.Items = append(data.Items, entry)
							}
						}
					}
				}
			}
			err := buildEntryListSection(mrt, cfg, sectionCfg, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func buildHeaderSection(mrt core.Maroto, cfg *utils.Config, content *utils.Content) error {
	headerFont := cfg.Fonts["header"]
	headerColor := utils.ResolveColor(headerFont.Color, cfg.Colors)
	headerSpacing := cfg.Spacing[cfg.SectionTemplates["header"].Spacing]

	mrt.AddRow(headerSpacing,
		text.NewCol(12, content.Personal.Name, props.Text{
			Family: headerFont.Family,
			Size:   headerFont.Size,
			Style:  utils.ResolveFontStyle(headerFont.Style),
			Color:  &headerColor,
			Align:  align.Left,
			Top:    5.0,
		}),
	)

	return nil
}

func buildContactSection(mrt core.Maroto, cfg *utils.Config, content *utils.Content) error {
	contactTemplate := cfg.SectionTemplates["contact"]
	contactSpacing := cfg.Spacing[contactTemplate.Spacing]

	// Process contact fields in groups of 3 per row (2x3 layout)
	for i := 0; i < len(content.ContactFields); i += 3 {
		var cols []core.Col

		// First contact field
		if i < len(content.ContactFields) {
			col := utils.CreateContactFieldCol(content.ContactFields[i], cfg, content)
			cols = append(cols, col)
		}

		// Second contact field if exists
		if i+1 < len(content.ContactFields) {
			col := utils.CreateContactFieldCol(content.ContactFields[i+1], cfg, content)
			cols = append(cols, col)
		}

		// Third contact field if exists
		if i+2 < len(content.ContactFields) {
			col := utils.CreateContactFieldCol(content.ContactFields[i+2], cfg, content)
			cols = append(cols, col)
		}

		if len(cols) > 0 {
			mrt.AddRow(contactSpacing, cols...)
		}
	}

	return nil
}

func buildSimpleListSection(mrt core.Maroto, cfg *utils.Config, sectionCfg utils.SectionConfig, sectionData interface{}) error {
	template := cfg.SectionTemplates[sectionCfg.Template]
	
	// Add section title with icon
	err := utils.AddSectionTitle(mrt, cfg, sectionCfg, template)
	if err != nil {
		return err
	}

	bodyFont := cfg.Fonts[template.Font]
	bodyColor := utils.ResolveColor(bodyFont.Color, cfg.Colors)
	bodySpacing := cfg.Spacing[template.Spacing]

	// Handle different data types
	switch data := sectionData.(type) {
	case map[string]interface{}:
		// Handle object with content field (summary)
		if content, ok := data["content"].(string); ok {
			mrt.AddRow(bodySpacing,
				text.NewCol(12, content, props.Text{
					Family: bodyFont.Family,
					Size:   bodyFont.Size,
					Style:  utils.ResolveFontStyle(bodyFont.Style),
					Color:  &bodyColor,
					Align:  align.Left,
				}),
			)
		}
	case []interface{}:
		// Handle array of strings (skills, certifications)
		var items []string
		for _, item := range data {
			if str, ok := item.(string); ok {
				items = append(items, str)
			}
		}
		
		if len(items) > 0 {
			itemsText := strings.Join(items, " | ")
			mrt.AddRow(bodySpacing,
				text.NewCol(12, itemsText, props.Text{
					Family: bodyFont.Family,
					Size:   bodyFont.Size,
					Style:  utils.ResolveFontStyle(bodyFont.Style),
					Color:  &bodyColor,
					Align:  align.Left,
				}),
			)
		}
	}

	return nil
}

func buildEntryListSection(mrt core.Maroto, cfg *utils.Config, sectionCfg utils.SectionConfig, data utils.SectionData) error {
	template := cfg.SectionTemplates[sectionCfg.Template]
	
	// Add section title with icon
	err := utils.AddSectionTitle(mrt, cfg, sectionCfg, template)
	if err != nil {
		return err
	}

	if data.Items == nil {
		return nil
	}

	bodyFont := cfg.Fonts[template.Font]
	emphasisFont := cfg.Fonts["emphasis"]
	bodyColor := utils.ResolveColor(bodyFont.Color, cfg.Colors)
	emphasisColor := utils.ResolveColor(emphasisFont.Color, cfg.Colors)
	itemSpacing := cfg.Spacing[template.ItemSpacing]

	for _, item := range data.Items {
		// Job/Education title and company/institution
		var titleLine string
		if item.Title != nil && item.Company != nil {
			titleLine = *item.Title + " - " + *item.Company
		} else if item.Degree != nil && item.Institution != nil {
			titleLine = *item.Degree + " - " + *item.Institution
		}

		if titleLine != "" {
			mrt.AddRow(itemSpacing,
				text.NewCol(12, titleLine, props.Text{
					Family: emphasisFont.Family,
					Size:   emphasisFont.Size,
					Style:  utils.ResolveFontStyle(emphasisFont.Style),
					Color:  &emphasisColor,
					Align:  align.Left,
				}),
			)
		}

		// Location and dates
		var locationLine string
		parts := []string{}
		if item.Location != nil {
			parts = append(parts, *item.Location)
		}
		if item.StartDate != nil || item.EndDate != nil {
			dateStr := ""
			if item.StartDate != nil {
				dateStr += *item.StartDate
			}
			dateStr += " - "
			if item.EndDate != nil {
				dateStr += *item.EndDate
			}
			parts = append(parts, dateStr)
		}
		
		if len(parts) > 0 {
			locationLine = strings.Join(parts, " | ")
			mrt.AddRow(itemSpacing-2,
				text.NewCol(12, locationLine, props.Text{
					Family: bodyFont.Family,
					Size:   bodyFont.Size - 1,
					Style:  fontstyle.Italic,
					Color:  &bodyColor,
					Align:  align.Left,
				}),
			)
		}

		// Description bullets
		if item.Description != nil {
			for _, bullet := range item.Description {
				mrt.AddRow(itemSpacing-2,
					text.NewCol(12, "- "+bullet, props.Text{
						Family: bodyFont.Family,
						Size:   bodyFont.Size - 1,
						Color:  &bodyColor,
						Align:  align.Left,
						Left:   5.0,
					}),
				)
			}
		}
	}

	return nil
}