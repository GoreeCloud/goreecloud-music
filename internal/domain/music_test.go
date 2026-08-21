package domain

import "testing"

func TestUserValidate(t *testing.T) {
	t.Parallel()

	valid := User{ID: "user-1", DisplayName: "Listener", Role: RoleUser}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid user: %v", err)
	}

	invalid := User{ID: "user-1", DisplayName: "Listener", Role: "owner"}
	if err := invalid.Validate(); err == nil {
		t.Fatal("expected unsupported role to fail validation")
	}
}

func TestLibraryValidate(t *testing.T) {
	t.Parallel()

	valid := Library{ID: "library-1", Name: "Shared Music", RootPath: "/music/shared", Visibility: LibraryShared}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid library: %v", err)
	}

	missingPath := Library{ID: "library-1", Name: "Shared Music", Visibility: LibraryShared}
	if err := missingPath.Validate(); err == nil {
		t.Fatal("expected missing root path to fail validation")
	}
}
