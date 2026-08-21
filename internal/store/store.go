package store

import (
	"context"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

type MusicStore interface {
	UserByID(ctx context.Context, userID string) (domain.User, error)
	LibrariesForUser(ctx context.Context, userID string) ([]domain.Library, error)
	LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error)
}
