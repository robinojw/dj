package state

import (
	"sort"
	"sync"
)

type ThreadStore struct {
	mu      sync.RWMutex
	threads map[string]*ThreadState
	order   []string
}

func NewThreadStore() *ThreadStore {
	return &ThreadStore{
		threads: make(map[string]*ThreadState),
		order:   make([]string, 0),
	}
}

func (store *ThreadStore) Add(id string, title string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.threads[id] = NewThreadState(id, title)
	store.order = append(store.order, id)
}

func (store *ThreadStore) AddWithParent(id string, title string, parentID string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	thread := NewThreadState(id, title)
	thread.ParentID = parentID
	store.threads[id] = thread
	store.order = append(store.order, id)
}

func (store *ThreadStore) Get(id string) (*ThreadState, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	thread, exists := store.threads[id]
	return thread, exists
}

func (store *ThreadStore) Delete(id string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	delete(store.threads, id)
	store.order = removeFromSlice(store.order, id)
}

func (store *ThreadStore) All() []*ThreadState {
	store.mu.RLock()
	defer store.mu.RUnlock()

	result := make([]*ThreadState, 0, len(store.order))
	for _, id := range store.order {
		if thread, exists := store.threads[id]; exists {
			result = append(result, thread)
		}
	}
	return result
}

func (store *ThreadStore) IDs() []string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	ids := make([]string, 0, len(store.threads))
	for id := range store.threads {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (store *ThreadStore) UpdateStatus(id string, status string, title string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	thread, exists := store.threads[id]
	if !exists {
		return
	}
	thread.Status = status
	if title != "" {
		thread.Title = title
	}
}

func (store *ThreadStore) UpdateActivity(id string, activity string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	thread, exists := store.threads[id]
	if !exists {
		return
	}
	thread.Activity = activity
}

func (store *ThreadStore) Children(parentID string) []*ThreadState {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var children []*ThreadState
	for _, id := range store.order {
		thread := store.threads[id]
		if thread.ParentID == parentID {
			children = append(children, thread)
		}
	}
	return children
}

func (store *ThreadStore) Roots() []*ThreadState {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var roots []*ThreadState
	for _, id := range store.order {
		thread := store.threads[id]
		if thread.ParentID == "" {
			roots = append(roots, thread)
		}
	}
	return roots
}

func removeFromSlice(slice []string, target string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != target {
			result = append(result, item)
		}
	}
	return result
}
