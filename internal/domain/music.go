package domain

import (
	"errors"
	"strings"
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type LibraryVisibility string

const (
	LibraryPrivate LibraryVisibility = "private"
	LibraryShared  LibraryVisibility = "shared"
)

type User struct {
	ID          string
	DisplayName string
	Role        Role
	CreatedAt   time.Time
}

type Library struct {
	ID         string
	Name       string
	RootPath   string
	Visibility LibraryVisibility
	CreatedAt  time.Time
}

type LibraryMembership struct {
	LibraryID string
	UserID    string
	CanRead   bool
	CanManage bool
}

type Artist struct {
	ID        string
	Name      string
	SortName  string
	CreatedAt time.Time
}

type Album struct {
	ID            string
	LibraryID     string
	Title         string
	AlbumArtistID string
	ReleaseYear   int
	CreatedAt     time.Time
}

type Track struct {
	ID          string
	LibraryID   string
	AlbumID     string
	Title       string
	TrackNumber int
	DiscNumber  int
	DurationMS  int64
	CreatedAt   time.Time
}

type TrackFile struct {
	ID         string
	TrackID    string
	Path       string
	Codec      string
	Bitrate    int
	BitDepth   int
	SampleRate int
	Channels   int
	SizeBytes  int64
	CreatedAt  time.Time
}

func (u User) Validate() error {
	if strings.TrimSpace(u.ID) == "" {
		return errors.New("user id is required")
	}
	if strings.TrimSpace(u.DisplayName) == "" {
		return errors.New("display name is required")
	}
	if u.Role != RoleUser && u.Role != RoleAdmin {
		return errors.New("unsupported user role")
	}
	return nil
}

func (l Library) Validate() error {
	if strings.TrimSpace(l.ID) == "" {
		return errors.New("library id is required")
	}
	if strings.TrimSpace(l.Name) == "" {
		return errors.New("library name is required")
	}
	if strings.TrimSpace(l.RootPath) == "" {
		return errors.New("library root path is required")
	}
	if l.Visibility != LibraryPrivate && l.Visibility != LibraryShared {
		return errors.New("unsupported library visibility")
	}
	return nil
}
