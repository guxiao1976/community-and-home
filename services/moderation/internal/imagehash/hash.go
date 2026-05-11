package imagehash

import (
	"image"
	"math/bits"
)

// ImageHasher provides perceptual image hashing with similarity detection
// against a set of known violation hashes.
type ImageHasher struct {
	hashSize         int
	similarThreshold int
	violationHashes  map[uint64]string // hash -> category
}

// NewImageHasher creates a new ImageHasher with the given hash bit-width
// (hashSize) and Hamming-distance threshold for similarity matching.
func NewImageHasher(hashSize, similarThreshold int) *ImageHasher {
	return &ImageHasher{
		hashSize:         hashSize,
		similarThreshold: similarThreshold,
		violationHashes:  make(map[uint64]string),
	}
}

// Hash computes a perceptual hash for the given image.
// This is a placeholder that returns 0, nil. A real implementation
// would resize the image to hashSize x hashSize, convert to
// grayscale, compute DCT, and take the median to produce the hash bits.
func (h *ImageHasher) Hash(_ image.Image) (uint64, error) {
	return 0, nil
}

// Distance returns the Hamming distance between two hashes (number of
// differing bits).
func (h *ImageHasher) Distance(h1, h2 uint64) int {
	return bits.OnesCount64(h1 ^ h2)
}

// LoadViolationHashes replaces the current set of known violation hashes
// with the provided map (hash value -> violation category).
func (h *ImageHasher) LoadViolationHashes(hashes map[uint64]string) {
	h.violationHashes = make(map[uint64]string, len(hashes))
	for k, v := range hashes {
		h.violationHashes[k] = v
	}
}

// FindSimilar checks the given hash against all loaded violation hashes.
// It returns the closest match within the similarThreshold, or
// (0, "", 0, false) if no similar hash is found.
func (h *ImageHasher) FindSimilar(hash uint64) (matchedHash uint64, category string, distance int, found bool) {
	bestDist := h.similarThreshold + 1
	for vh, cat := range h.violationHashes {
		d := h.Distance(hash, vh)
		if d < bestDist {
			bestDist = d
			matchedHash = vh
			category = cat
		}
	}
	if bestDist <= h.similarThreshold {
		return matchedHash, category, bestDist, true
	}
	return 0, "", 0, false
}
