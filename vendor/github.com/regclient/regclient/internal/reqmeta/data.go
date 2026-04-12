// Package reqmeta provides metadata on requests for prioritizing with a pqueue.
package reqmeta

type Data struct {
	Kind Kind
	Size int64
}

type Kind int

const (
	Unknown Kind = iota
	Head
	Manifest
	Query
	Blob
)

const (
	smallLimit = 4194304 // 4MiB
	largePct   = 0.9     // anything above 90% of largest queued entry size is large
)

func DataNext(queued, active []*Data) int {
	if len(queued) == 0 {
		return -1
	}
	// After removing one small entry, split remaining requests 50/50 between large and old (truncated int division always rounds down).
	// If len active = 2, this function returns the 3rd entry (+1), minus 1 for the small, divide by 2 to split with old = goal of 1.
	largeGoal := len(active) / 2
	largeI := 0
	var largeSize int64
	if largeGoal > 0 {
		//  find the largest queued blob requests
		for i, cur := range queued {
			if cur.Kind == Blob && cur.Size > largeSize {
				largeI = i
				largeSize = cur.Size
			}
		}
	}
	largeCutoff := int64(float64(largeSize) * 0.9)
	// count active requests by type
	small := 0
	large := 0
	old := 0
	for _, cur := range active {
		if cur.Kind != Blob && cur.Size <= smallLimit {
			small++
		} else if cur.Kind == Blob && largeSize > 0 && cur.Size >= largeCutoff {
			large++
		} else {
			old++
		}
	}
	// if there is at least one active, and none are small, return the best small entry if available.
	if len(active) > 0 && small == 0 {
		var sizeI int64
		bestI := -1
		kindI := Unknown
		for i, cur := range queued {
			// the small search skips blobs and large requests
			if cur.Kind == Blob || cur.Size > smallLimit {
				continue
			}
			// the best small entry is the:
			//  - first one found if no other matches
			//  - one with a better Kind (Head > Manifest > Query)
			//  - one with the same kind but smaller request
			if bestI < 0 ||
				(cur.Kind != Unknown && (kindI == Unknown || cur.Kind < kindI)) ||
				(cur.Kind == kindI && cur.Size > 0 && (cur.Size < sizeI || sizeI <= 0)) {
				bestI = i
				kindI = cur.Kind
				sizeI = cur.Size
			}
		}
		if bestI >= 0 {
			return bestI
		}
	}
	// Prefer the biggest of these blobs to minimize the size of the last running blob.
	if largeGoal > 0 && large < largeGoal && largeSize > 0 {
		return largeI
	}
	// enough small and large, or none available, so return the oldest queued entry to avoid starvation.
	return 0
}
