/*
Copyright 2011 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Author: lode.vandevenne@gmail.com (Lode Vandevenne)
Author: jyrki.alakuijala@gmail.com (Jyrki Alakuijala)
*/

package zopfli

// litLen: Contains the literal symbol or length value.
// dists: Indicates the distance, or 0 to indicate that there is no distance and
//   litLens contains a literal instead of a length.
type lz77Pair struct {
	// Lit or len.
	litLen uint16

	// If 0: indicates literal in corresponding litLens,
	// if > 0: length in corresponding litLens, this is the distance.
	dist uint16
}

// Stores lit/length and dist pairs for LZ77.
type LZ77Store []lz77Pair

// Some state information for compressing a block.
// This is currently a bit under-used (with mainly only the longest match cache),
// but is kept for easy future expansion.
type BlockState struct {
	options *Options
	block []byte

	// The start (inclusive) and end (not inclusive) of the current block.
	blockStart, blockEnd int

	// Cache for length/distance pairs found so far.
	lmc longestMatchCache
}

type lz77Counts struct {
	litLen, dist []uint
}

// Gets a score of the length given the distance. Typically, the score of the
// length is the length itself, but if the distance is very long, decrease the
// score of the length a bit to make up for the fact that long distances use
// large amounts of extra bits.

// This is not an accurate score, it is a heuristic only for the greedy LZ77
// implementation. More accurate cost models are employed later. Making this
// heuristic more accurate may hurt rather than improve compression.

// The two direct uses of this heuristic are:
// -avoid using a length of 3 in combination with a long distance. This only has
//  an effect if length == 3.
// -make a slightly better choice between the two options of the lazy matching.

// Indirectly, this affects:
// -the block split points if the default of block splitting first is used, in a
//  rather unpredictable way
// -the first zopfli run, so it affects the chance of the first run being closer
//  to the optimal output
func (pair lz77Pair) lengthScore() uint16 {
	// At 1024, the distance uses 9+ extra bits and this seems to be the
	// sweet spot on tested files.
	if pair.dist > 1024 {
		return pair.litLen - 1
	}
	return pair.litLen
}

// Verifies if length and dist are indeed valid, only used for assertion.
func (pair lz77Pair) Verify(data []byte, pos int) {

	// TODO(lode): make this only run in a debug compile, it's for assert only.

	dataSize := len(data)
	if pos+int(pair.litLen) > dataSize {
		panic("overrun")
	}
	for i := 0; i < int(pair.litLen); i++ {
		if data[pos-int(pair.dist)+i] != data[pos+i] {
			panic("mismatch")
		}
	}
}

// Finds how long the match of scan and match is. Can be used to find how many
// bytes starting from scan, and from match, are equal. Returns the last byte
// after scan, which is still equal to the corresponding byte after match.
// scan is the position to compare
// match is the earlier position to compare.
// end is the last possible position, beyond which to stop looking.
func getMatch(slice []byte, scan, match, end int) int {
	for scan < end && slice[scan] == slice[match] {
		scan++
		match++
	}

	return scan
}

func NewBlockState(options *Options, in []byte, inStart, inEnd int) (s BlockState) {
	s.options = options
	s.block = in
	s.blockStart = inStart
	s.blockEnd = inEnd
	if LONGEST_MATCH_CACHE {
		blockSize := inEnd - inStart
		s.lmc = newCache(blockSize)
	}
	return s
}

// Gets distance, length and sublen values from the cache if possible.
// Returns 1 if it got the values from the cache, 0 if not.
// Updates the limit value to a smaller one if possible with more limited
// information from the cache.
func (s *BlockState) tryGetFromLongestMatchCache(pos int, limit *uint16, sublen []uint16) (pair lz77Pair, ok bool) {
	// The LMC cache starts at the beginning of the block rather than the
	// beginning of the whole array.
	lmcPos := pos - s.blockStart

	// Length > 0 and dist 0 is invalid combination, which indicates on
	// purpose that this cache value is not filled in yet.
	cacheAvailable := s.lmc.active && (s.lmc.store[lmcPos].litLen == 0 || s.lmc.store[lmcPos].dist != 0)
	var maxSublen uint16
	if cacheAvailable && sublen != nil {
		maxSublen = s.lmc.maxCachedSublen(lmcPos)
	}
	limitOkForCache := cacheAvailable && (*limit == MAX_MATCH || s.lmc.store[lmcPos].litLen <= *limit || (sublen != nil && maxSublen >= *limit))
	if s.lmc.active && limitOkForCache && cacheAvailable {
		if sublen == nil || s.lmc.store[lmcPos].litLen <= s.lmc.maxCachedSublen(lmcPos) {
			pair.litLen = s.lmc.store[lmcPos].litLen
			if pair.litLen > *limit {
				pair.litLen = *limit
			}
			if sublen != nil {
				s.lmc.cacheToSublen(lmcPos, pair.litLen, sublen)
				pair.dist = sublen[pair.litLen]
				if *limit == MAX_MATCH && pair.litLen >= MIN_MATCH {
					if pair.dist != s.lmc.store[lmcPos].dist {
						panic("bad sublen")
					}
				}
			} else {
				pair.dist = s.lmc.store[lmcPos].dist
			}
			return pair, true
		}
		// Can't use much of the cache, since the "sublens" need to
		// be calculated, but at least we already know when to stop.
		*limit = s.lmc.store[lmcPos].litLen
	}

	return pair, false
}

// Stores the found sublen, distance and length in the longest match cache, if
// possible.
func (s *BlockState) storeInLongestMatchCache(pos int, limit uint16,
	sublen []uint16, pair lz77Pair) {
	if !s.lmc.active {
		return
	}

	// The LMC cache starts at the beginning of the block rather than the
	// beginning of the whole array.
	lmcPos := pos - s.blockStart
	lmcPair := s.lmc.store[lmcPos]

	// Length > 0 and dist 0 is invalid combination, which indicates on purpose
	// that this cache value is not filled in yet.
	cacheAvailable := lmcPair.litLen == 0 || lmcPair.dist != 0
	if cacheAvailable {
		return
	}

	if limit != MAX_MATCH || sublen == nil {
		return
	}

	if lmcPair.litLen != 1 || lmcPair.dist != 0 {
		panic("overrun")
	}
	if pair.litLen < MIN_MATCH {
		lmcPair = lz77Pair{}
	} else {
		lmcPair = pair
	}
	s.lmc.store[lmcPos] = lmcPair
	if lmcPair.litLen == 1 && lmcPair.dist == 0 {
		panic("cached invalid combination")
	}
	s.lmc.sublenToCache(sublen, lmcPos, pair.litLen)
}

// Finds the longest match (length and corresponding distance) for LZ77
// compression.
// Even when not using "sublen", it can be more efficient to provide an array,
// because only then the caching is used.
//
// slice: the data
//
// pos: position in the data to find the match for
//
// size: size of the data
//
// limit: limit length to maximum this value (default should be 258). This allows
// finding a shorter dist for that length (= less extra bits). Must be
// in the range [MIN_MATCH, MAX_MATCH].
//
// sublen: output array of 259 elements, or null. Has, for each length, the
// smallest distance required to reach this length. Only 256 of its 259 values
// are used, the first 3 are ignored (the shortest length is 3. It is purely
// for convenience that the array is made 3 longer).
func (s *BlockState) findLongestMatch(h *hash, slice []byte, pos, size int, limit uint16, sublen []uint16) lz77Pair {
	hPos := uint16(pos & WINDOW_MASK)
	bestPair := lz77Pair{1, 0}
	chainCounter := MAX_CHAIN_HITS // For quitting early.
	hPrev := h.prev

	if LONGEST_MATCH_CACHE {
		pair, ok := s.tryGetFromLongestMatchCache(pos, &limit, sublen)
		if ok {
			/*
				if pos+int(pair.litLen) > size {
					panic("overrun")
				}
			*/
			return pair
		}
	}

	if limit > MAX_MATCH {
		panic("limit is too large")
	} else if limit < MIN_MATCH {
		panic("limit is too small")
	}
	if pos >= size {
		panic("overrun")
	}

	if size < pos+MIN_MATCH {
		// The rest of the code assumes there are at least MIN_MATCH
		// bytes to try.
		return lz77Pair{}
	}

	if pos+int(limit) > size {
		limit = uint16(size - pos)
	}
	arrayEnd := pos + int(limit)

	if h.val >= 65536 {
		panic("hash value too large")
	}

	pp := uint16(h.head[h.val]) // During the whole loop, p == h.prev[pp].
	p := hPrev[pp]

	if pp != hPos {
		panic("invalid pp")
	}

	var dist int // Not uint16 on purpose.
	if p < pp {
		dist = int(pp - p)
	} else {
		dist = WINDOW_SIZE + int(pp) - int(p)
	}

	// Go through all distances.
	same0 := h.same[hPos]
	scanned := slice[pos]
	for dist < WINDOW_SIZE {
		scan := pos
		match := pos - dist

		// Testing the byte at position bestLength first, goes slightly faster.
		var currentLength uint16
		bestLitLen := bestPair.litLen
		bestPos := pos + int(bestLitLen)
		bestMatch := match + int(bestLitLen)
		if bestPos >= size || slice[bestPos] == slice[bestMatch] {
			if HASH_SAME {
				if same0 > 2 && scanned == slice[match] {
					same1 := h.same[match&WINDOW_MASK]
					var same uint16
					if same0 < same1 {
						same = same0
					} else {
						same = same1
					}
					if same > limit {
						same = limit
					}
					scan += int(same)
					match += int(same)
				}
			}
			scan = getMatch(slice, scan, match, arrayEnd)
			currentLength = uint16(scan - pos) // The found length.

			if currentLength > bestLitLen {
				if sublen != nil {
					for j := bestLitLen + 1; j <= currentLength; j++ {
						sublen[j] = uint16(dist)
					}
				}
				bestPair = lz77Pair{currentLength, uint16(dist)}
				if currentLength >= limit {
					break
				}
			}
		}

		if HASH_SAME_HASH {
			// Switch to the other hash once this will be more efficient.
			if bestPair.litLen >= same0 && h.val2 == h.hashVal2[p] {
				// Now use the hash that encodes the length and first byte.
				hPrev = h.prev2
			}
		}

		pp = p
		p = hPrev[p]
		if p == pp {
			// Uninited prev value.
			break
		}

		if p >= pp {
			dist += WINDOW_SIZE
		}
		dist += int(pp) - int(p)

		if MAX_CHAIN_HITS < WINDOW_SIZE {
			chainCounter--
			if chainCounter <= 0 {
				break
			}
		}
	}

	if LONGEST_MATCH_CACHE {
		s.storeInLongestMatchCache(pos, limit, sublen, bestPair)
	}

	if bestPair.litLen > limit {
		panic("overrun")
	}

	if pos+int(bestPair.litLen) > size {
		panic("overrun")
	}
	return bestPair
}

// Does LZ77 using an algorithm similar to gzip, with lazy matching, rather than
// with the slow but better "squeeze" implementation.
// The result is placed in the LZ77Store.
// If inStart is larger than 0, it uses values before inStart as starting
// dictionary.
func (s *BlockState) LZ77Greedy(inStart, inEnd int) (store LZ77Store) {
	var windowStart int
	if inStart > WINDOW_SIZE {
		windowStart = inStart - WINDOW_SIZE
	}

	// Lazy matching.
	var prevPair lz77Pair
	var matchAvailable bool

	if inStart == inEnd {
		return
	}

	h := newHash(s.block[windowStart], s.block[windowStart+1])
	for i := windowStart; i < inStart; i++ {
		h.update(s.block, i, inEnd)
	}

	dummySublen := make([]uint16, 259)
	for i := inStart; i < inEnd; i++ {
		h.update(s.block, i, inEnd)

		pair := s.findLongestMatch(&h, s.block, i, inEnd, MAX_MATCH, dummySublen[:])
		lengthScore := pair.lengthScore()

		if LAZY_MATCHING {
			// Lazy matching.
			prevLengthScore := prevPair.lengthScore()
			if matchAvailable {
				matchAvailable = false
				if lengthScore > prevLengthScore+1 {
					store = append(store, lz77Pair{uint16(s.block[i-1]), 0})
					if lengthScore >= MIN_MATCH && pair.litLen < MAX_MATCH {
						matchAvailable = true
						prevPair = pair
						continue
					}
				} else {
					// Add previous to output.
					pair = prevPair
					lengthScore = prevLengthScore
					// Add to output.
					pair.Verify(s.block, i-1)
					store = append(store, pair)
					for j := uint16(2); j < pair.litLen; j++ {
						if i >= inEnd {
							panic("overrun")
						}
						i++
						h.update(s.block, i, inEnd)
					}
					continue
				}
			} else if lengthScore >= MIN_MATCH && pair.litLen < MAX_MATCH {
				matchAvailable = true
				prevPair = pair
				continue
			}
			// End of lazy matching.
		}

		// Add to output.
		if lengthScore >= MIN_MATCH {
			pair.Verify(s.block, i)
			store = append(store, pair)
		} else {
			pair.litLen = 1
			store = append(store, lz77Pair{uint16(s.block[i]), 0})
		}
		for j := uint16(1); j < pair.litLen; j++ {
			if i >= inEnd {
				panic("overrun")
			}
			i++
			h.update(s.block, i, inEnd)
		}
	}
	return store
}

// Counts the number of literal, length and distance symbols in the given lz77
// arrays.
// litLens: lz77 lit/lengths
// dists: ll77 distances
// start: where to begin counting in litLens and dists
// end: where to stop counting in litLens and dists (not inclusive)
// llCount: count of each lit/len symbol, must have size 288 (see deflate
//     standard)
// dCount: count of each dist symbol, must have size 32 (see deflate standard)
func (store LZ77Store) lz77Counts() (counts lz77Counts) {
	counts.litLen = make([]uint, 288)
	counts.dist = make([]uint, 32)
	end := len(store)
	for i := 0; i < end; i++ {
		pair := store[i]
		if pair.dist == 0 {
			counts.litLen[pair.litLen]++
		} else {
			counts.litLen[pair.lengthSymbol()]++
			counts.dist[pair.distSymbol()]++
		}
	}

	counts.litLen[256] = 1 // End symbol.
	return counts
}
