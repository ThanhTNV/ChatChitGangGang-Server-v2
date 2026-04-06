package dbmigrate

import "embed"

//go:embed migrations/*.sql
var sqlFiles embed.FS
