// Package logger provides a minimal structured logger for cronitor-local.
//
// Log lines are written in a human-readable format:
//
//	2024-01-15T10:00:00Z [INFO] job=backup-db command started
//	2024-01-15T10:00:05Z [ERROR] job=backup-db exit status 1
//
// Callers choose a minimum Level at construction time; messages below that
// level are silently discarded. Use LevelDebug during development and
// LevelInfo (or higher) in production.
package logger
