// Package jobfilter provides tag-based and name-based filtering for jobs.
package jobfilter

import "strings"

// Filter holds compiled filter criteria for selecting a subset of jobs.
type Filter struct {
	names map[string]struct{}
	tags  map[string]struct{}
}

// New returns a Filter that matches jobs by any of the given names or tags.
// Passing empty slices creates a filter that matches all jobs.
func New(names, tags []string) *Filter {
	nm := make(map[string]struct{}, len(names))
	for _, n := range names {
		nm[strings.TrimSpace(n)] = struct{}{}
	}
	tm := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		tm[strings.TrimSpace(t)] = struct{}{}
	}
	return &Filter{names: nm, tags: tm}
}

// MatchName reports whether the given job name is accepted by the filter.
// If no name criteria were provided every name is accepted.
func (f *Filter) MatchName(name string) bool {
	if len(f.names) == 0 {
		return true
	}
	_, ok := f.names[name]
	return ok
}

// MatchTags reports whether at least one of the supplied job tags is accepted
// by the filter. If no tag criteria were provided every set of tags is accepted.
func (f *Filter) MatchTags(jobTags []string) bool {
	if len(f.tags) == 0 {
		return true
	}
	for _, t := range jobTags {
		if _, ok := f.tags[t]; ok {
			return true
		}
	}
	return false
}

// Match reports whether a job with the given name and tags passes the filter.
// A job passes when it satisfies both the name criterion and the tag criterion.
func (f *Filter) Match(name string, jobTags []string) bool {
	return f.MatchName(name) && f.MatchTags(jobTags)
}

// IsEmpty reports whether the filter has no criteria, i.e. it matches
// everything.
func (f *Filter) IsEmpty() bool {
	return len(f.names) == 0 && len(f.tags) == 0
}
