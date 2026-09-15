package storage

import "testing"

func TestCurrentSchemaValid(t *testing.T) {
	schema := CurrentSchema()
	if schema.Version != CurrentVersion {
		t.Fatalf("version = %d, want %d", schema.Version, CurrentVersion)
	}
	if err := schema.Validate(); err != nil {
		t.Fatalf("current schema invalid: %v", err)
	}
}

func TestCurrentSchemaDoesNotEmbedMediaOrSecrets(t *testing.T) {
	for _, entity := range CurrentSchema().Entities {
		if entity.ContainsMedia {
			t.Fatalf("entity %q embeds media bytes", entity.Name)
		}
		if entity.ContainsSecret {
			t.Fatalf("entity %q embeds reusable secrets", entity.Name)
		}
	}
}

func TestSchemaValidationRejectsDuplicateEntity(t *testing.T) {
	schema := CurrentSchema()
	schema.Entities = append(schema.Entities, schema.Entities[0])
	if err := schema.Validate(); err == nil {
		t.Fatal("expected duplicate entity validation failure")
	}
}
