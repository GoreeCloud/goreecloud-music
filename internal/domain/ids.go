package domain

import (
	"fmt"
	"strings"
	"unicode"
)

type RecordingID string
type ReleaseID string
type SourceItemID string
type PlayableAssetID string
type QueueItemID string
type ProfileID string
type LibraryID string

func validateID(kind, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s must not be empty", kind)
	}
	if len(value) > 128 {
		return fmt.Errorf("%s exceeds 128 characters", kind)
	}
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_.:", r) {
			continue
		}
		return fmt.Errorf("%s contains unsupported character %q", kind, r)
	}
	return nil
}

func NewRecordingID(value string) (RecordingID, error) {
	if err := validateID("recording id", value); err != nil {
		return "", err
	}
	return RecordingID(strings.TrimSpace(value)), nil
}
func NewReleaseID(value string) (ReleaseID, error) {
	if err := validateID("release id", value); err != nil {
		return "", err
	}
	return ReleaseID(strings.TrimSpace(value)), nil
}
func NewSourceItemID(value string) (SourceItemID, error) {
	if err := validateID("source item id", value); err != nil {
		return "", err
	}
	return SourceItemID(strings.TrimSpace(value)), nil
}
func NewPlayableAssetID(value string) (PlayableAssetID, error) {
	if err := validateID("playable asset id", value); err != nil {
		return "", err
	}
	return PlayableAssetID(strings.TrimSpace(value)), nil
}
func NewQueueItemID(value string) (QueueItemID, error) {
	if err := validateID("queue item id", value); err != nil {
		return "", err
	}
	return QueueItemID(strings.TrimSpace(value)), nil
}
func NewProfileID(value string) (ProfileID, error) {
	if err := validateID("profile id", value); err != nil {
		return "", err
	}
	return ProfileID(strings.TrimSpace(value)), nil
}
func NewLibraryID(value string) (LibraryID, error) {
	if err := validateID("library id", value); err != nil {
		return "", err
	}
	return LibraryID(strings.TrimSpace(value)), nil
}
