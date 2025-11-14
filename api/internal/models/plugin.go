package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Repository represents a plugin repository
type Repository struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Type        string    `json:"type"` // git, http
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CatalogPlugin represents a plugin available in the catalog
type CatalogPlugin struct {
	ID             int             `json:"id"`
	RepositoryID   int             `json:"repositoryId"`
	Name           string          `json:"name"`
	Version        string          `json:"version"`
	DisplayName    string          `json:"displayName"`
	Description    string          `json:"description"`
	Category       string          `json:"category"`
	PluginType     string          `json:"pluginType"` // extension, webhook, api, ui, theme
	IconURL        string          `json:"iconUrl"`
	Manifest       PluginManifest  `json:"manifest"`
	Tags           []string        `json:"tags"`
	InstallCount   int             `json:"installCount"`
	AvgRating      float64         `json:"avgRating"`
	RatingCount    int             `json:"ratingCount"`
	Repository     Repository      `json:"repository"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// InstalledPlugin represents a plugin installed in the system
type InstalledPlugin struct {
	ID              int             `json:"id"`
	CatalogPluginID *int            `json:"catalogPluginId,omitempty"`
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Enabled         bool            `json:"enabled"`
	Config          json.RawMessage `json:"config,omitempty"`
	InstalledBy     string          `json:"installedBy"`
	InstalledAt     time.Time       `json:"installedAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`

	// Populated from catalog if available
	DisplayName     string          `json:"displayName,omitempty"`
	Description     string          `json:"description,omitempty"`
	PluginType      string          `json:"pluginType,omitempty"`
	IconURL         string          `json:"iconUrl,omitempty"`
	Manifest        *PluginManifest `json:"manifest,omitempty"`
}

// PluginManifest contains plugin metadata and configuration
type PluginManifest struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	DisplayName     string                 `json:"displayName"`
	Description     string                 `json:"description"`
	Author          string                 `json:"author"`
	Homepage        string                 `json:"homepage,omitempty"`
	Repository      string                 `json:"repository,omitempty"`
	License         string                 `json:"license,omitempty"`
	Type            string                 `json:"type"` // extension, webhook, api, ui, theme
	Category        string                 `json:"category,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Icon            string                 `json:"icon,omitempty"`

	// Requirements
	Requirements    PluginRequirements     `json:"requirements,omitempty"`

	// Entry points
	Entrypoints     PluginEntrypoints      `json:"entrypoints,omitempty"`

	// Configuration schema
	ConfigSchema    map[string]interface{} `json:"configSchema,omitempty"`
	DefaultConfig   map[string]interface{} `json:"defaultConfig,omitempty"`

	// Permissions
	Permissions     []string               `json:"permissions,omitempty"`

	// Dependencies
	Dependencies    map[string]string      `json:"dependencies,omitempty"`
}

// PluginRequirements specifies plugin requirements
type PluginRequirements struct {
	StreamSpaceVersion string            `json:"streamspaceVersion,omitempty"` // e.g., ">=0.2.0"
	MinimumVersion     string            `json:"minimumVersion,omitempty"`
	MaximumVersion     string            `json:"maximumVersion,omitempty"`
	Plugins            []string          `json:"plugins,omitempty"` // Required plugins
}

// PluginEntrypoints defines plugin entry points
type PluginEntrypoints struct {
	Main      string `json:"main,omitempty"`      // Main entry point
	UI        string `json:"ui,omitempty"`        // UI component entry point
	API       string `json:"api,omitempty"`       // API routes entry point
	Webhook   string `json:"webhook,omitempty"`   // Webhook handler
	CLI       string `json:"cli,omitempty"`       // CLI command entry point
}

// PluginVersion represents a version of a plugin
type PluginVersion struct {
	ID        int             `json:"id"`
	PluginID  int             `json:"pluginId"`
	Version   string          `json:"version"`
	Changelog string          `json:"changelog,omitempty"`
	Manifest  PluginManifest  `json:"manifest"`
	CreatedAt time.Time       `json:"createdAt"`
}

// PluginRating represents a user's rating for a plugin
type PluginRating struct {
	ID        int       `json:"id"`
	PluginID  int       `json:"pluginId"`
	UserID    string    `json:"userId"`
	Rating    int       `json:"rating"` // 1-5
	Review    string    `json:"review,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PluginStats represents usage statistics for a plugin
type PluginStats struct {
	PluginID        int        `json:"pluginId"`
	ViewCount       int        `json:"viewCount"`
	InstallCount    int        `json:"installCount"`
	LastViewedAt    *time.Time `json:"lastViewedAt,omitempty"`
	LastInstalledAt *time.Time `json:"lastInstalledAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// Scan implements sql.Scanner for PluginManifest
func (m *PluginManifest) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for PluginManifest
func (m PluginManifest) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// InstallPluginRequest represents a request to install a plugin
type InstallPluginRequest struct {
	PluginID int             `json:"pluginId"` // From catalog
	Config   json.RawMessage `json:"config,omitempty"`
}

// UpdatePluginRequest represents a request to update a plugin
type UpdatePluginRequest struct {
	Enabled *bool           `json:"enabled,omitempty"`
	Config  json.RawMessage `json:"config,omitempty"`
}

// RatePluginRequest represents a request to rate a plugin
type RatePluginRequest struct {
	Rating int    `json:"rating"` // 1-5
	Review string `json:"review,omitempty"`
}
