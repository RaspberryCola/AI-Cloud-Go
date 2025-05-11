package utils

import (
	"strings"
)

// ReplaceKnowledgePlaceholder replaces the {{Knowledge}} placeholder in the prompt
func ReplaceKnowledgePlaceholder(prompt string, knowledge string) string {
	return strings.Replace(prompt, "{{Knowledge}}", knowledge, -1)
}

// FormatRetrievalResults formats the retrieval results for insertion into the prompt
func FormatRetrievalResults(results []map[string]interface{}) string {
	var builder strings.Builder

	for i, result := range results {
		if i > 0 {
			builder.WriteString("\n\n")
		}

		// Extract content and metadata
		if content, ok := result["content"].(string); ok {
			builder.WriteString(content)
		}

		// Add source if available
		if source, ok := result["source"].(string); ok {
			builder.WriteString("\nSource: " + source)
		}
	}

	return builder.String()
}
