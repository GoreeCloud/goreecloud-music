package storage

import (
	"context"
	"fmt"
	"strings"
)

// ChangeKind identifies an engine-neutral schema operation.
type ChangeKind string

const (
	ChangeCreateEntity ChangeKind = "create-entity"
)

// Change is an engine-neutral migration instruction. Concrete storage backends
// translate these declarations into their own transactional operations.
type Change struct {
	Kind   ChangeKind
	Entity string
	Detail string
}

// Migration is one forward-only durable-state schema transition.
type Migration struct {
	ID      string
	From    Version
	To      Version
	Changes []Change
}

var migrations = []Migration{
	{
		ID:   "0001-core-application-state",
		From: UninitializedVersion,
		To:   1,
		Changes: []Change{
			{Kind: ChangeCreateEntity, Entity: "schema_metadata", Detail: "schema version and migration metadata"},
			{Kind: ChangeCreateEntity, Entity: "profiles", Detail: "Music profile application state; authentication credentials remain external"},
			{Kind: ChangeCreateEntity, Entity: "libraries", Detail: "library identity, ownership, and source-root metadata"},
			{Kind: ChangeCreateEntity, Entity: "recordings", Detail: "canonical recording identity and metadata"},
			{Kind: ChangeCreateEntity, Entity: "releases", Detail: "release and edition identity"},
			{Kind: ChangeCreateEntity, Entity: "source_items", Detail: "source-provider identity mapping"},
			{Kind: ChangeCreateEntity, Entity: "playable_assets", Detail: "encoded asset metadata only; original media bytes stay outside application database state"},
			{Kind: ChangeCreateEntity, Entity: "queues", Detail: "profile-owned queue identity and revision"},
			{Kind: ChangeCreateEntity, Entity: "queue_items", Detail: "ordered recording/source intent and routing state"},
			{Kind: ChangeCreateEntity, Entity: "favorites", Detail: "profile-scoped favorite state"},
			{Kind: ChangeCreateEntity, Entity: "ratings", Detail: "profile-scoped rating state"},
			{Kind: ChangeCreateEntity, Entity: "play_history", Detail: "profile-scoped listening history with privacy and retention controls"},
		},
	},
	{
		ID:   "0002-library-memberships",
		From: 1,
		To:   2,
		Changes: []Change{
			{Kind: ChangeCreateEntity, Entity: "library_memberships", Detail: "per-profile library authorization with explicit roles; concrete backends must preserve existing library-owner access when materializing membership state"},
		},
	},
	{
		ID:   "0003-library-files",
		From: 2,
		To:   CurrentVersion,
		Changes: []Change{
			{Kind: ChangeCreateEntity, Entity: "library_files", Detail: "per-library filesystem observations for source-preserving incremental scan and missing-file reconciliation"},
		},
	},
}

// Backend is the minimum transactional boundary needed by the migration runner.
type Backend interface {
	SchemaVersion(context.Context) (Version, error)
	ApplyMigration(context.Context, Migration) error
}

// Catalog returns a defensive copy of the ordered migration catalog.
func Catalog() []Migration {
	return cloneMigrations(migrations)
}

// ValidateCatalog ensures the migration history is contiguous, forward-only,
// uniquely identified, internally valid, and reaches CurrentVersion.
func ValidateCatalog(catalog []Migration) error {
	if len(catalog) == 0 {
		return fmt.Errorf("migration catalog must not be empty")
	}

	schema := CurrentSchema()
	if err := schema.Validate(); err != nil {
		return fmt.Errorf("current schema invalid: %w", err)
	}
	validEntities := make(map[string]struct{}, len(schema.Entities))
	for _, entity := range schema.Entities {
		validEntities[entity.Name] = struct{}{}
	}

	expected := UninitializedVersion
	ids := make(map[string]struct{}, len(catalog))
	created := make(map[string]struct{}, len(validEntities))
	for i, migration := range catalog {
		if strings.TrimSpace(migration.ID) == "" {
			return fmt.Errorf("migration %d must define an id", i)
		}
		if _, ok := ids[migration.ID]; ok {
			return fmt.Errorf("duplicate migration id %q", migration.ID)
		}
		ids[migration.ID] = struct{}{}
		if migration.From != expected {
			return fmt.Errorf("migration %q starts at %d, expected %d", migration.ID, migration.From, expected)
		}
		if migration.To <= migration.From {
			return fmt.Errorf("migration %q must move forward", migration.ID)
		}
		if len(migration.Changes) == 0 {
			return fmt.Errorf("migration %q must declare at least one change", migration.ID)
		}
		for j, change := range migration.Changes {
			if change.Kind != ChangeCreateEntity {
				return fmt.Errorf("migration %q change %d has unsupported kind %q", migration.ID, j, change.Kind)
			}
			entity := strings.TrimSpace(change.Entity)
			if entity == "" {
				return fmt.Errorf("migration %q change %d must name an entity", migration.ID, j)
			}
			if _, ok := validEntities[entity]; !ok {
				return fmt.Errorf("migration %q change %d references unknown entity %q", migration.ID, j, entity)
			}
			if _, ok := created[entity]; ok {
				return fmt.Errorf("entity %q is created more than once", entity)
			}
			created[entity] = struct{}{}
			if strings.TrimSpace(change.Detail) == "" {
				return fmt.Errorf("migration %q change %d must define detail", migration.ID, j)
			}
		}
		expected = migration.To
	}
	if expected != CurrentVersion {
		return fmt.Errorf("migration catalog ends at %d, current schema is %d", expected, CurrentVersion)
	}
	for entity := range validEntities {
		if _, ok := created[entity]; !ok {
			return fmt.Errorf("migration catalog does not create current entity %q", entity)
		}
	}
	return nil
}

// Plan returns the forward-only migrations required to reach target.
func Plan(from, target Version) ([]Migration, error) {
	if target > CurrentVersion {
		return nil, fmt.Errorf("target schema %d is newer than supported current schema %d", target, CurrentVersion)
	}
	if from > target {
		return nil, fmt.Errorf("downgrade from schema %d to %d is not supported", from, target)
	}
	if from == target {
		return nil, nil
	}
	if err := ValidateCatalog(migrations); err != nil {
		return nil, err
	}

	plan := make([]Migration, 0, len(migrations))
	version := from
	for _, migration := range migrations {
		if migration.From < version {
			continue
		}
		if migration.From != version {
			return nil, fmt.Errorf("no migration path from schema %d", version)
		}
		if migration.To > target {
			break
		}
		plan = append(plan, migration)
		version = migration.To
		if version == target {
			return cloneMigrations(plan), nil
		}
	}
	return nil, fmt.Errorf("no migration path from schema %d to %d", from, target)
}

func cloneMigrations(in []Migration) []Migration {
	out := make([]Migration, len(in))
	for i, migration := range in {
		out[i] = migration
		out[i].Changes = append([]Change(nil), migration.Changes...)
	}
	return out
}

// Migrate executes a validated forward migration plan and verifies the backend
// reached the requested target version. Backends are responsible for atomicity
// of each ApplyMigration call and must update their stored schema version only
// after a successful migration.
func Migrate(ctx context.Context, backend Backend, target Version) error {
	if backend == nil {
		return fmt.Errorf("storage backend must not be nil")
	}
	current, err := backend.SchemaVersion(ctx)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	plan, err := Plan(current, target)
	if err != nil {
		return err
	}
	for _, migration := range plan {
		if err := backend.ApplyMigration(ctx, migration); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.ID, err)
		}
	}
	final, err := backend.SchemaVersion(ctx)
	if err != nil {
		return fmt.Errorf("read final schema version: %w", err)
	}
	if final != target {
		return fmt.Errorf("backend schema version is %d after migration, expected %d", final, target)
	}
	return nil
}
