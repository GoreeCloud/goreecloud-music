package sqlitestore

import (
	"context"
	"fmt"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

// RecentlyAddedRecording is one canonical recording currently visible to a
// profile, ordered by when it entered the Music application-state library.
type RecentlyAddedRecording struct {
	RecordingID domain.RecordingID
	LibraryID   domain.LibraryID
	Title       string
	Artist      string
	AddedAt     time.Time
}

// RecentlyAddedForProfile returns the newest canonical recordings that the
// profile may currently read. Revoked or absent library membership suppresses
// those recordings from this view without deleting their underlying library
// state.
func (b *Backend) RecentlyAddedForProfile(ctx context.Context, profileID domain.ProfileID, limit int) ([]RecentlyAddedRecording, error) {
	profile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return nil, err
	}
	if err := validateProfileStateLimit(limit); err != nil {
		return nil, err
	}

	rows, err := b.db.QueryContext(ctx, `SELECT r.recording_id, r.library_id, r.title, r.artist, r.added_at
		FROM recordings r
		JOIN library_memberships m ON m.library_id = r.library_id
		WHERE m.profile_id = ?
		ORDER BY r.added_at DESC, r.recording_id
		LIMIT ?`, string(profile), limit)
	if err != nil {
		return nil, fmt.Errorf("list recently added recordings: %w", err)
	}
	defer rows.Close()

	items := make([]RecentlyAddedRecording, 0)
	for rows.Next() {
		var recordingRaw, libraryRaw, title, artist, addedRaw string
		if err := rows.Scan(&recordingRaw, &libraryRaw, &title, &artist, &addedRaw); err != nil {
			return nil, fmt.Errorf("scan recently added recording: %w", err)
		}
		addedAt, err := time.Parse(time.RFC3339Nano, addedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse recording added_at: %w", err)
		}
		items = append(items, RecentlyAddedRecording{
			RecordingID: domain.RecordingID(recordingRaw),
			LibraryID:   domain.LibraryID(libraryRaw),
			Title:       title,
			Artist:      artist,
			AddedAt:     addedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recently added recordings: %w", err)
	}
	return items, nil
}
