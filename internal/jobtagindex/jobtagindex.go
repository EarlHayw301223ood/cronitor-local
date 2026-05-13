// Package jobtagindex maintains a reverse index from tag to job names,
// enabling fast lookup of all jobs associated with a given tag.
package jobtagindex

import "sync"

// Index maps tags to the set of job names that carry that tag.
type Index struct {
	mu   sync.RWMutex
	data map[string]map[string]struct{} // tag -> set of job names
}

// New returns an empty Index.
func New() *Index {
	return &Index{
		data: make(map[string]map[string]struct{}),
	}
}

// Register associates jobName with each of the provided tags.
// Calling Register again for the same job replaces its previous tag set.
func (idx *Index) Register(jobName string, tags []string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove stale associations for this job.
	for tag, jobs := range idx.data {
		delete(jobs, jobName)
		if len(jobs) == 0 {
			delete(idx.data, tag)
		}
	}

	for _, tag := range tags {
		if _, ok := idx.data[tag]; !ok {
			idx.data[tag] = make(map[string]struct{})
		}
		idx.data[tag][jobName] = struct{}{}
	}
}

// JobsForTag returns a sorted slice of job names associated with tag.
// Returns nil if the tag is unknown.
func (idx *Index) JobsForTag(tag string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	jobs, ok := idx.data[tag]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(jobs))
	for name := range jobs {
		out = append(out, name)
	}
	return out
}

// Tags returns all tags currently tracked by the index.
func (idx *Index) Tags() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	out := make([]string, 0, len(idx.data))
	for tag := range idx.data {
		out = append(out, tag)
	}
	return out
}

// Deregister removes all tag associations for jobName.
func (idx *Index) Deregister(jobName string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for tag, jobs := range idx.data {
		delete(jobs, jobName)
		if len(jobs) == 0 {
			delete(idx.data, tag)
		}
	}
}
