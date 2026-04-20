package rtk

import "hash/fnv"

// dedupRing is a fixed-size circular buffer that tracks recently seen line hashes
// for duplicate detection. It uses FNV-1a hashing for speed.
type dedupRing struct {
	hashes []uint64
	size   int
	pos    int
	count  int
}

func newDedupRing(size int) *dedupRing {
	if size <= 0 {
		size = 32
	}
	return &dedupRing{
		hashes: make([]uint64, size),
		size:   size,
	}
}

// seen returns true if the line was already in the ring buffer.
// Either way, it adds the line's hash to the ring.
func (r *dedupRing) seen(line string) bool {
	h := hashLine(line)

	// Check if this hash exists in the ring.
	limit := r.count
	if limit > r.size {
		limit = r.size
	}
	for i := 0; i < limit; i++ {
		if r.hashes[i] == h {
			return true
		}
	}

	// Add to ring.
	r.hashes[r.pos] = h
	r.pos = (r.pos + 1) % r.size
	r.count++
	return false
}

func hashLine(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
