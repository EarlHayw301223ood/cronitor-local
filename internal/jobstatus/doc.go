// Package jobstatus tracks the last known execution status for each
// monitored job. It provides a lightweight in-memory store that records
// whether a job succeeded or failed, along with the timestamp of the
// most recent state change.
//
// The store is safe for concurrent use and is intended to be queried
// by the health-check server and the alert manager.
package jobstatus
