package work

// Waiting is how many jobs wait, for every member together.
func (q *Queue) Waiting() int {
	if q == nil {
		return 0
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.waiting)
}
