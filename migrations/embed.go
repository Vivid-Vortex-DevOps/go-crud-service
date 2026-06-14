package migrations

import "embed"

// FS contains all SQL migration files embedded at build time.
// The embed directive must live in this package (same directory as the .sql files)
// because Go's embed does not support '..' in patterns.
//
//go:embed *.sql
var FS embed.FS
