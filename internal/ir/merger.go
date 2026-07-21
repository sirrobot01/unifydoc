package ir

import (
	"fmt"
	"sort"
)

// Merger handles merging multiple IR instances into a unified representation
type Merger struct {
	irs []*IR
}

// NewMerger creates a new Merger instance
func NewMerger() *Merger {
	return &Merger{
		irs: make([]*IR, 0),
	}
}

// Add adds an IR to the merger
func (m *Merger) Add(ir *IR) {
	if ir != nil {
		m.irs = append(m.irs, ir)
	}
}

// Merge combines all IRs into a unified structure
func (m *Merger) Merge() (*MergedIR, error) {
	if len(m.irs) == 0 {
		return nil, fmt.Errorf("no IRs to merge")
	}

	merged := &MergedIR{
		Protocols:    make([]*IR, 0),
		AllServers:   make([]Server, 0),
		AllResources: make([]Resource, 0),
		AllTypes:     make([]TypeDef, 0),
		AllSecurity:  make([]SecurityScheme, 0),
		Navigation:   make([]NavigationItem, 0),
		SearchIndex:  make([]SearchEntry, 0),
	}

	// Add all protocols
	merged.Protocols = append(merged.Protocols, m.irs...)

	// Build navigation structure
	for _, ir := range m.irs {
		navItem := NavigationItem{
			Protocol: ir.Protocol,
			Title:    ir.Title,
			Count:    len(ir.Resources),
			Children: make([]NavigationChild, 0),
		}

		// Group resources by tags or create default group
		tagGroups := m.groupResourcesByTags(ir.Resources)

		for tag, resources := range tagGroups {
			child := NavigationChild{
				Name:  tag,
				Count: len(resources),
			}
			navItem.Children = append(navItem.Children, child)
		}

		merged.Navigation = append(merged.Navigation, navItem)

		// Aggregate servers
		merged.AllServers = append(merged.AllServers, ir.Servers...)

		// Aggregate resources
		merged.AllResources = append(merged.AllResources, ir.Resources...)

		// Aggregate types
		merged.AllTypes = append(merged.AllTypes, ir.Types...)

		// Aggregate security schemes
		merged.AllSecurity = append(merged.AllSecurity, ir.Security...)

		// Build search index entries
		for _, resource := range ir.Resources {
			entry := SearchEntry{
				Protocol:    ir.Protocol,
				Name:        resource.Name,
				Path:        resource.Path,
				Method:      resource.Method,
				Description: resource.Description,
				Tags:        resource.Tags,
			}
			merged.SearchIndex = append(merged.SearchIndex, entry)
		}
	}

	// Sort navigation by protocol name
	sort.Slice(merged.Navigation, func(i, j int) bool {
		return merged.Navigation[i].Protocol < merged.Navigation[j].Protocol
	})

	return merged, nil
}

// groupResourcesByTags groups resources by their tags
func (m *Merger) groupResourcesByTags(resources []Resource) map[string][]Resource {
	groups := make(map[string][]Resource)

	for _, resource := range resources {
		if len(resource.Tags) == 0 {
			// Default group
			groups["General"] = append(groups["General"], resource)
		} else {
			for _, tag := range resource.Tags {
				groups[tag] = append(groups[tag], resource)
			}
		}
	}

	return groups
}

// MergedIR represents the unified IR from all protocols
type MergedIR struct {
	Protocols    []*IR            `json:"protocols"`
	AllServers   []Server         `json:"all_servers"`
	AllResources []Resource       `json:"all_resources"`
	AllTypes     []TypeDef        `json:"all_types"`
	AllSecurity  []SecurityScheme `json:"all_security"`
	Navigation   []NavigationItem `json:"navigation"`
	SearchIndex  []SearchEntry    `json:"search_index"`
}

// NavigationItem represents a protocol in the navigation
type NavigationItem struct {
	Protocol string            `json:"protocol"`
	Title    string            `json:"title"`
	Count    int               `json:"count"`
	Children []NavigationChild `json:"children"`
}

// NavigationChild represents a grouped section within a protocol
type NavigationChild struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SearchEntry represents a searchable entry
type SearchEntry struct {
	Protocol    string   `json:"protocol"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}
