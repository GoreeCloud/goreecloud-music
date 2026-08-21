package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

var ErrNotFound = errors.New("not found")

type PostgresStore struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) UserByID(ctx context.Context, userID string) (domain.User, error) {
	const query = `SELECT id::text, display_name, role, created_at FROM users WHERE id::text = $1`
	var user domain.User
	if err := s.db.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.DisplayName, &user.Role, &user.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s *PostgresStore) LibrariesForUser(ctx context.Context, userID string) ([]domain.Library, error) {
	const query = `
		SELECT l.id::text, l.name, l.root_path, l.visibility, l.created_at
		FROM libraries l
		JOIN library_memberships m ON m.library_id = l.id
		WHERE m.user_id::text = $1 AND m.can_read = true
		ORDER BY l.name, l.id`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libraries []domain.Library
	for rows.Next() {
		var library domain.Library
		if err := rows.Scan(&library.ID, &library.Name, &library.RootPath, &library.Visibility, &library.CreatedAt); err != nil {
			return nil, err
		}
		libraries = append(libraries, library)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return libraries, nil
}

func (s *PostgresStore) LibraryMembership(ctx context.Context, libraryID, userID string) (domain.LibraryMembership, error) {
	const query = `
		SELECT library_id::text, user_id::text, can_read, can_manage
		FROM library_memberships
		WHERE library_id::text = $1 AND user_id::text = $2`
	var membership domain.LibraryMembership
	if err := s.db.QueryRowContext(ctx, query, libraryID, userID).Scan(
		&membership.LibraryID,
		&membership.UserID,
		&membership.CanRead,
		&membership.CanManage,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.LibraryMembership{}, ErrNotFound
		}
		return domain.LibraryMembership{}, err
	}
	return membership, nil
}
