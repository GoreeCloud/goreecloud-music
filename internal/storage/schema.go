package storage

import (
	"fmt"
	"sort"
	"strings"
)

// Version identifies an application-state schema version. It is independent
// from the GoreeCloud Music software release version.
type Version uint32

const (
	UninitializedVersion Version = 0
	CurrentVersion       Version = 3
)

// OwnershipScope documents the authorization boundary that owns a durable
// application-state entity. It is intentionally engine-neutral.
type OwnershipScope string

const (
	OwnershipSystem  OwnershipScope = "system"
	OwnershipProfile OwnershipScope = "profile"
	OwnershipLibrary OwnershipScope = "library"
)

// EntityDescriptor defines a durable application-state entity without choosing
// a concrete database engine or storage encoding.
type EntityDescriptor struct {
	Name           string
	PrimaryKey     string
	Ownership      OwnershipScope
	Purpose        string
	ContainsMedia  bool
	ContainsSecret bool
}

// Schema describes the durable application-state model expected by a backend.
type Schema struct {
	Version  Version
	Entities []EntityDescriptor
}

var currentEntities = []EntityDescriptor{
	{Name: "schema_metadata", PrimaryKey: "key", Ownership: OwnershipSystem, Purpose: "schema version and migration metadata"},
	{Name: "profiles", PrimaryKey: "profile_id", Ownership: OwnershipProfile, Purpose: "Music profile application state without authentication credentials"},
	{Name: "libraries", PrimaryKey: "library_id", Ownership: OwnershipLibrary, Purpose: "library identity, ownership, and source-root metadata"},
	{Name: "library_memberships", PrimaryKey: "library_id+profile_id", Ownership: OwnershipLibrary, Purpose: "per-profile library authorization and role state"},
	{Name: "library_files", PrimaryKey: "file_id", Ownership: OwnershipLibrary, Purpose: "source-preserving filesystem observations used for incremental library reconciliation"},
	{Name: "recordings", PrimaryKey: "recording_id", Ownership: OwnershipLibrary, Purpose: "canonical recording identity and library metadata"},
	{Name: "releases", PrimaryKey: "release_id", Ownership: OwnershipLibrary, Purpose: "release and edition identity"},
	{Name: "source_items", PrimaryKey: "source_item_id", Ownership: OwnershipLibrary, Purpose: "provider-specific source identity mappings"},
	{Name: "playable_assets", PrimaryKey: "playable_asset_id", Ownership: OwnershipLibrary, Purpose: "encoded asset metadata, quality, path reference, and availability facts"},
	{Name: "queues", PrimaryKey: "queue_id", Ownership: OwnershipProfile, Purpose: "profile-owned durable queue identity and revision state"},
	{Name: "queue_items", PrimaryKey: "queue_item_id", Ownership: OwnershipProfile, Purpose: "queue ordering, recording/source intent, selected route, and route reason"},
	{Name: "favorites", PrimaryKey: "favorite_id", Ownership: OwnershipProfile, Purpose: "profile-scoped favorite state"},
	{Name: "ratings", PrimaryKey: "rating_id", Ownership: OwnershipProfile, Purpose: "profile-scoped rating state"},
	{Name: "play_history", PrimaryKey: "history_id", Ownership: OwnershipProfile, Purpose: "profile-scoped listening history subject to privacy and retention controls"},
}

// CurrentSchema returns a defensive copy of the current engine-neutral storage model.
func CurrentSchema() Schema {
	entities := make([]EntityDescriptor, len(currentEntities))
	copy(entities, currentEntities)
	return Schema{Version: CurrentVersion, Entities: entities}
}

// Validate verifies that a schema is internally consistent and preserves the
// GoreeCloud Music boundary that original media bytes are not database state.
func (s Schema) Validate() error {
	if s.Version == UninitializedVersion {
		return fmt.Errorf("schema version must be initialized")
	}
	if len(s.Entities) == 0 {
		return fmt.Errorf("schema must contain at least one entity")
	}

	seen := make(map[string]struct{}, len(s.Entities))
	for _, entity := range s.Entities {
		name := strings.TrimSpace(entity.Name)
		if name == "" {
			return fmt.Errorf("entity name must not be empty")
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("duplicate entity %q", name)
		}
		seen[name] = struct{}{}
		if strings.TrimSpace(entity.PrimaryKey) == "" {
			return fmt.Errorf("entity %q must define a primary key", name)
		}
		switch entity.Ownership {
		case OwnershipSystem, OwnershipProfile, OwnershipLibrary:
		default:
			return fmt.Errorf("entity %q has unsupported ownership scope %q", name, entity.Ownership)
		}
		if strings.TrimSpace(entity.Purpose) == "" {
			return fmt.Errorf("entity %q must define a purpose", name)
		}
		if entity.ContainsMedia {
			return fmt.Errorf("entity %q must not embed original media bytes in application database state", name)
		}
		if entity.ContainsSecret {
			return fmt.Errorf("entity %q must not embed reusable credentials or secrets", name)
		}
	}

	required := []string{"schema_metadata", "profiles", "libraries", "library_memberships", "library_files", "recordings", "releases", "source_items", "playable_assets", "queues", "queue_items"}
	for _, name := range required {
		if _, ok := seen[name]; !ok {
			return fmt.Errorf("schema missing required entity %q", name)
		}
	}
	return nil
}

// EntityNames returns sorted entity names for deterministic diagnostics and tests.
func (s Schema) EntityNames() []string {
	names := make([]string, 0, len(s.Entities))
	for _, entity := range s.Entities {
		names = append(names, entity.Name)
	}
	sort.Strings(names)
	return names
}
