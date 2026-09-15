package sqlitestore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/GoreeCloud/goreecloud-music/internal/domain"
)

const maxProfileStateResults = 100

// FavoriteState is profile-owned favorite state for an authorized recording.
type FavoriteState struct {
	RecordingID domain.RecordingID
	CreatedAt   time.Time
}

// RatingState is profile-owned rating state for an authorized recording.
type RatingState struct {
	RecordingID domain.RecordingID
	Rating      int
	UpdatedAt   time.Time
}

// PlayHistoryEntry is one profile-owned listening-history event. Library
// authorization is re-evaluated when history is read so revoked libraries do
// not remain visible through Music history surfaces.
type PlayHistoryEntry struct {
	HistoryID   string
	RecordingID domain.RecordingID
	PlayedAt    time.Time
	PlayedMS    int64
}

// SetFavorite creates or clears profile-owned favorite state. Read permission
// to the recording's library is required for either mutation.
func (b *Backend) SetFavorite(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID, favorite bool) error {
	profile, recording, err := validateProfileRecording(profileID, recordingID)
	if err != nil {
		return err
	}
	if err := b.requireRecordingPermission(ctx, profile, recording, LibraryPermissionRead); err != nil {
		return err
	}
	if !favorite {
		_, err := b.db.ExecContext(ctx, `DELETE FROM favorites WHERE profile_id = ? AND recording_id = ?`, string(profile), string(recording))
		if err != nil {
			return fmt.Errorf("clear favorite: %w", err)
		}
		return nil
	}

	favoriteID, err := newProfileStateID("favorite")
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = b.db.ExecContext(ctx, `INSERT INTO favorites(favorite_id, profile_id, recording_id, created_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(profile_id, recording_id) DO NOTHING`, favoriteID, string(profile), string(recording), now)
	if err != nil {
		return fmt.Errorf("set favorite: %w", err)
	}
	return nil
}

// FavoritesForProfile returns only favorite recordings that the profile may
// currently read. Revoked library access suppresses stale favorite rows.
func (b *Backend) FavoritesForProfile(ctx context.Context, profileID domain.ProfileID, limit int) ([]FavoriteState, error) {
	profile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return nil, err
	}
	if err := validateProfileStateLimit(limit); err != nil {
		return nil, err
	}
	rows, err := b.db.QueryContext(ctx, `SELECT f.recording_id, f.created_at
		FROM favorites f
		JOIN recordings r ON r.recording_id = f.recording_id
		JOIN library_memberships m ON m.library_id = r.library_id AND m.profile_id = f.profile_id
		WHERE f.profile_id = ?
		ORDER BY f.created_at DESC, f.recording_id
		LIMIT ?`, string(profile), limit)
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	states := make([]FavoriteState, 0)
	for rows.Next() {
		var recordingRaw, createdRaw string
		if err := rows.Scan(&recordingRaw, &createdRaw); err != nil {
			return nil, fmt.Errorf("scan favorite: %w", err)
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
		if err != nil {
			return nil, fmt.Errorf("parse favorite created_at: %w", err)
		}
		states = append(states, FavoriteState{RecordingID: domain.RecordingID(recordingRaw), CreatedAt: createdAt})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorites: %w", err)
	}
	return states, nil
}

// SetRating creates or updates a profile-owned 0-100 rating for an authorized
// recording. Read permission is sufficient because the rating is private
// profile state and does not modify the library object.
func (b *Backend) SetRating(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID, rating int) error {
	profile, recording, err := validateProfileRecording(profileID, recordingID)
	if err != nil {
		return err
	}
	if rating < 0 || rating > 100 {
		return fmt.Errorf("rating must be between 0 and 100")
	}
	if err := b.requireRecordingPermission(ctx, profile, recording, LibraryPermissionRead); err != nil {
		return err
	}

	ratingID, err := newProfileStateID("rating")
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = b.db.ExecContext(ctx, `INSERT INTO ratings(rating_id, profile_id, recording_id, rating, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, recording_id) DO UPDATE SET rating = excluded.rating, updated_at = excluded.updated_at`, ratingID, string(profile), string(recording), rating, now)
	if err != nil {
		return fmt.Errorf("set rating: %w", err)
	}
	return nil
}

// ClearRating removes profile-owned rating state for an authorized recording.
func (b *Backend) ClearRating(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID) error {
	profile, recording, err := validateProfileRecording(profileID, recordingID)
	if err != nil {
		return err
	}
	if err := b.requireRecordingPermission(ctx, profile, recording, LibraryPermissionRead); err != nil {
		return err
	}
	if _, err := b.db.ExecContext(ctx, `DELETE FROM ratings WHERE profile_id = ? AND recording_id = ?`, string(profile), string(recording)); err != nil {
		return fmt.Errorf("clear rating: %w", err)
	}
	return nil
}

// RatingForRecording returns a profile rating only while that profile retains
// read authorization to the recording's library.
func (b *Backend) RatingForRecording(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID) (RatingState, bool, error) {
	profile, recording, err := validateProfileRecording(profileID, recordingID)
	if err != nil {
		return RatingState{}, false, err
	}
	if err := b.requireRecordingPermission(ctx, profile, recording, LibraryPermissionRead); err != nil {
		return RatingState{}, false, err
	}
	var value int
	var updatedRaw string
	err = b.db.QueryRowContext(ctx, `SELECT rating, updated_at FROM ratings WHERE profile_id = ? AND recording_id = ?`, string(profile), string(recording)).Scan(&value, &updatedRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return RatingState{}, false, nil
	}
	if err != nil {
		return RatingState{}, false, fmt.Errorf("read rating: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedRaw)
	if err != nil {
		return RatingState{}, false, fmt.Errorf("parse rating updated_at: %w", err)
	}
	return RatingState{RecordingID: recording, Rating: value, UpdatedAt: updatedAt}, true, nil
}

// RecordPlay appends one profile-owned listening-history event for an
// authorized recording. This records application history only; operational
// telemetry and security audit records remain separate concerns.
func (b *Backend) RecordPlay(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID, playedAt time.Time, playedMS int64) error {
	profile, recording, err := validateProfileRecording(profileID, recordingID)
	if err != nil {
		return err
	}
	if playedAt.IsZero() {
		return fmt.Errorf("played_at must not be zero")
	}
	if playedMS < 0 {
		return fmt.Errorf("played_ms must not be negative")
	}
	if err := b.requireRecordingPermission(ctx, profile, recording, LibraryPermissionRead); err != nil {
		return err
	}
	historyID, err := newProfileStateID("history")
	if err != nil {
		return err
	}
	_, err = b.db.ExecContext(ctx, `INSERT INTO play_history(history_id, profile_id, recording_id, played_at, played_ms)
		VALUES(?, ?, ?, ?, ?)`, historyID, string(profile), string(recording), playedAt.UTC().Format(time.RFC3339Nano), playedMS)
	if err != nil {
		return fmt.Errorf("record play history: %w", err)
	}
	return nil
}

// RecentPlays returns profile-owned history for recordings that the profile may
// currently read. Authorization loss suppresses stale history from normal Music
// surfaces without rewriting or leaking another profile's history.
func (b *Backend) RecentPlays(ctx context.Context, profileID domain.ProfileID, limit int) ([]PlayHistoryEntry, error) {
	profile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return nil, err
	}
	if err := validateProfileStateLimit(limit); err != nil {
		return nil, err
	}
	rows, err := b.db.QueryContext(ctx, `SELECT h.history_id, h.recording_id, h.played_at, h.played_ms
		FROM play_history h
		JOIN recordings r ON r.recording_id = h.recording_id
		JOIN library_memberships m ON m.library_id = r.library_id AND m.profile_id = h.profile_id
		WHERE h.profile_id = ?
		ORDER BY h.played_at DESC, h.history_id DESC
		LIMIT ?`, string(profile), limit)
	if err != nil {
		return nil, fmt.Errorf("list recent plays: %w", err)
	}
	defer rows.Close()

	entries := make([]PlayHistoryEntry, 0)
	for rows.Next() {
		var historyID, recordingRaw, playedRaw string
		var playedMS int64
		if err := rows.Scan(&historyID, &recordingRaw, &playedRaw, &playedMS); err != nil {
			return nil, fmt.Errorf("scan play history: %w", err)
		}
		playedAt, err := time.Parse(time.RFC3339Nano, playedRaw)
		if err != nil {
			return nil, fmt.Errorf("parse played_at: %w", err)
		}
		entries = append(entries, PlayHistoryEntry{
			HistoryID:   historyID,
			RecordingID: domain.RecordingID(recordingRaw),
			PlayedAt:    playedAt,
			PlayedMS:    playedMS,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent plays: %w", err)
	}
	return entries, nil
}

func (b *Backend) requireRecordingPermission(ctx context.Context, profileID domain.ProfileID, recordingID domain.RecordingID, required LibraryPermission) error {
	var actualRaw string
	err := b.db.QueryRowContext(ctx, `SELECT m.permission
		FROM recordings r
		JOIN library_memberships m ON m.library_id = r.library_id
		WHERE r.recording_id = ? AND m.profile_id = ?`, string(recordingID), string(profileID)).Scan(&actualRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("recording is not accessible to profile")
	}
	if err != nil {
		return fmt.Errorf("resolve recording authorization: %w", err)
	}
	actual := LibraryPermission(actualRaw)
	if err := validatePermission(actual); err != nil {
		return fmt.Errorf("stored recording permission invalid: %w", err)
	}
	if permissionRank(actual) < permissionRank(required) {
		return fmt.Errorf("recording is not accessible to profile")
	}
	return nil
}

func validateProfileRecording(profileID domain.ProfileID, recordingID domain.RecordingID) (domain.ProfileID, domain.RecordingID, error) {
	profile, err := domain.NewProfileID(string(profileID))
	if err != nil {
		return "", "", err
	}
	recording, err := domain.NewRecordingID(string(recordingID))
	if err != nil {
		return "", "", err
	}
	return profile, recording, nil
}

func validateProfileStateLimit(limit int) error {
	if limit < 1 || limit > maxProfileStateResults {
		return fmt.Errorf("limit must be between 1 and %d", maxProfileStateResults)
	}
	return nil
}

func newProfileStateID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	return prefix + ":" + hex.EncodeToString(raw[:]), nil
}
