package migrations

import "embed"

// Files contains the versioned SQL migrations shipped with GoreeCloud Music.
//
//go:embed *.sql
var Files embed.FS
