package generator

import (
	"strings"

	"github.com/sirrobot01/unifydoc/internal/ir"
)

// SearchIndexer builds a search index for client-side search
type SearchIndexer struct {
	entries []SearchIndexEntry
}

// SearchIndexEntry represents a searchable entry
type SearchIndexEntry struct {
	ID          string   `json:"id"`
	Protocol    string   `json:"protocol"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	SearchText  string   `json:"search_text"`
}

// NewSearchIndexer creates a new search indexer
func NewSearchIndexer() *SearchIndexer {
	return &SearchIndexer{
		entries: make([]SearchIndexEntry, 0),
	}
}

// Build builds the search index from merged IR
func (s *SearchIndexer) Build(merged *ir.MergedIR) {
	s.entries = make([]SearchIndexEntry, 0)

	for _, protocol := range merged.Protocols {
		for i, resource := range protocol.Resources {
			entry := SearchIndexEntry{
				ID:          generateID(protocol.Protocol, i),
				Protocol:    protocol.Protocol,
				Name:        resource.Name,
				Path:        resource.Path,
				Method:      resource.Method,
				Description: resource.Description,
				Tags:        resource.Tags,
			}

			// Build search text
			searchTextParts := []string{
				resource.Name,
				resource.Path,
				resource.Method,
				resource.Description,
				protocol.Protocol,
				protocol.Title,
			}

			// Add tags
			searchTextParts = append(searchTextParts, resource.Tags...)

			// Add parameter names
			for _, param := range resource.Parameters {
				searchTextParts = append(searchTextParts, param.Name, param.Description)
			}

			entry.SearchText = strings.ToLower(strings.Join(searchTextParts, " "))

			s.entries = append(s.entries, entry)
		}
	}
}

// GetIndex returns the search index
func (s *SearchIndexer) GetIndex() []SearchIndexEntry {
	return s.entries
}

// generateID generates a unique ID for a resource
func generateID(protocol string, index int) string {
	return strings.ToLower(protocol) + "-" + string(rune(index))
}
