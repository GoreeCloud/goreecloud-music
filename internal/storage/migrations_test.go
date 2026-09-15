package storage

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type fakeBackend struct {
	version Version
	applied []string
	failID  string
}

func (b *fakeBackend) SchemaVersion(context.Context) (Version, error) {
	return b.version, nil
}

func (b *fakeBackend) ApplyMigration(_ context.Context, migration Migration) error {
	if migration.ID == b.failID {
		return errors.New("injected migration failure")
	}
	if b.version != migration.From {
		return errors.New("unexpected migration start version")
	}
	b.applied = append(b.applied, migration.ID)
	b.version = migration.To
	return nil
}

func TestCatalogValid(t *testing.T) {
	if err := ValidateCatalog(Catalog()); err != nil {
		t.Fatalf("catalog invalid: %v", err)
	}
}

func TestPlanFromUninitializedToCurrent(t *testing.T) {
	plan, err := Plan(UninitializedVersion, CurrentVersion)
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	ids := make([]string, len(plan))
	for i, migration := range plan {
		ids[i] = migration.ID
	}
	if want := []string{"0001-core-application-state", "0002-library-memberships"}; !reflect.DeepEqual(ids, want) {
		t.Fatalf("plan ids = %v, want %v", ids, want)
	}
}

func TestPlanRejectsDowngrade(t *testing.T) {
	if _, err := Plan(CurrentVersion, UninitializedVersion); err == nil {
		t.Fatal("expected downgrade to fail")
	}
}

func TestMigrateAdvancesBackend(t *testing.T) {
	backend := &fakeBackend{version: UninitializedVersion}
	if err := Migrate(context.Background(), backend, CurrentVersion); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if backend.version != CurrentVersion {
		t.Fatalf("version = %d, want %d", backend.version, CurrentVersion)
	}
	if want := []string{"0001-core-application-state", "0002-library-memberships"}; !reflect.DeepEqual(backend.applied, want) {
		t.Fatalf("applied = %v, want %v", backend.applied, want)
	}
}

func TestMigrateFailsClosedOnBackendError(t *testing.T) {
	backend := &fakeBackend{version: UninitializedVersion, failID: "0001-core-application-state"}
	if err := Migrate(context.Background(), backend, CurrentVersion); err == nil {
		t.Fatal("expected migration error")
	}
	if backend.version != UninitializedVersion {
		t.Fatalf("version advanced after failed migration: %d", backend.version)
	}
}
