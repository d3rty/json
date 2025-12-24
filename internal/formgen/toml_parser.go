package formgen

import (
	"bufio"
	"strings"
)

// TOMLComments holds a mapping of TOML keys to their comment descriptions.
type TOMLComments map[string]string

// ParseTOMLComments reads TOML content and extracts comments above each field.
// It builds a map from field paths (e.g., "Bool.FromStrings.CaseInsensitive") to their descriptions.
func ParseTOMLComments(content string) (TOMLComments, error) {
	comments := make(TOMLComments)
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentSection []string
	var pendingComment strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			pendingComment.Reset()
			continue
		}

		// Handle section headers like [Bool], [Bool.FromStrings]
		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.Trim(trimmed, "[]")
			currentSection = strings.Split(sectionName, ".")
			pendingComment.Reset()
			continue
		}

		// Handle comments
		if strings.HasPrefix(trimmed, "#") {
			commentText := strings.TrimPrefix(trimmed, "#")
			commentText = strings.TrimSpace(commentText)

			// Skip comments that are just section descriptions (they start with the section name)
			if len(currentSection) > 0 {
				sectionPrefix := strings.Join(currentSection, ".")
				if strings.HasPrefix(commentText, sectionPrefix) {
					continue
				}
			}

			if pendingComment.Len() > 0 {
				pendingComment.WriteString(" ")
			}
			pendingComment.WriteString(commentText)
			continue
		}

		// Handle field assignments like "CaseInsensitive = true"
		if strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				fieldName := strings.TrimSpace(parts[0])

				// Build the full path
				var fullPath string
				if len(currentSection) > 0 {
					fullPath = strings.Join(currentSection, ".") + "." + fieldName
				} else {
					fullPath = fieldName
				}

				// Store the comment if we have one
				if pendingComment.Len() > 0 {
					comments[fullPath] = pendingComment.String()
					pendingComment.Reset()
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetDescription returns the description for a given field path.
// It tries multiple variations to handle nested structures.
func (tc TOMLComments) GetDescription(fieldPath string) string {
	// Try exact match first
	if desc, ok := tc[fieldPath]; ok {
		return desc
	}

	// Try case variations
	for key, desc := range tc {
		if strings.EqualFold(key, fieldPath) {
			return desc
		}
	}

	return ""
}
