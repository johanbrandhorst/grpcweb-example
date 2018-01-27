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

import (
	"fmt"
	"math"
	"os"
)

type symbolStats struct {
	// The literal and length symbols.
	litLens []float64
	// The 32 unique dist symbols, not the 32768 possible dists.
	dists []float64

	// Length of each lit/len symbol in bits.
	llSymbols []float64
	// Length of each dist symbol in bits.
	dSymbols []float64
}

// Adds the bit lengths.
func addWeightedFreqs(stats1 symbolStats, w1 float64,
	stats2 symbolStats, w2 float64) (result symbolStats) {
	result.litLens = make([]float64, 288)
	result.dists = make([]float64, 32)
	for i := 0; i < 288; i++ {
		litlen1 := stats1.litLens[i] * w1
		litlen2 := stats2.litLens[i] * w2
		result.litLens[i] = litlen1 + litlen2
	}
	for i := 0; i < 32; i++ {
		dist1 := stats1.dists[i] * w1
		dist2 := stats2.dists[i] * w2
		result.dists[i] = dist1 + dist2
	}
	result.litLens[256] = 1 // End symbol.
	return result
}

type ranState struct {
	m_w, m_z uint32
}

func newRanState() (state ranState) {
	state.m_w = 1
	state.m_z = 2
	return state
}

/* Get random number: "Multiply-With-Carry" generator of G. Marsaglia */
func (state ranState) ran() uint32 {
	state.m_z = 36969*(state.m_z&65535) + (state.m_z >> 16)
	state.m_w = 18000*(state.m_w&65535) + (state.m_w >> 16)
	return (state.m_z << 16) + state.m_w // 32-bit result.
}

func (state ranState) randomizeFreqs(freqs []float64) {
	n := uint32(len(freqs))
	for i := uint32(0); i < n; i++ {
		if (state.ran()>>4)%3 == 0 {
			freqs[i] = freqs[state.ran()%n]
		}
	}
}

func (stats symbolStats) randomizeFreqs(state ranState) {
	state.randomizeFreqs(stats.litLens)
	state.randomizeFreqs(stats.dists)
	stats.litLens[256] = 1 // End symbol.
}

// Function that calculates a cost based on a model for the given LZ77 symbol.
// litlen: means literal symbol if dist is 0, length otherwise.
type costModelFun func(pair lz77Pair, context interface{}) float64

// Cost model which should exactly match fixed tree.
// type: costModelFun
func costFixed(pair lz77Pair, unused interface{}) float64 {
	if pair.dist == 0 {
		if pair.litLen <= 143 {
			return 8
		}
		return 9
	}
	dBits := pair.distExtraBits()
	lBits := pair.lengthExtraBits()
	lSym := pair.lengthSymbol()
	cost := float64(5) // Every dist symbol has length 5.
	if lSym <= 279 {
		cost += 7
	} else {
		cost += 8
	}
	return cost + float64(dBits+lBits)
}

// Cost model based on symbol statistics.
// type: costModelFun
func costStat(pair lz77Pair, context interface{}) float64 {
	stats := context.(symbolStats)
	if pair.dist == 0 {
		return stats.llSymbols[pair.litLen]
	}
	lSym := pair.lengthSymbol()
	lBits := pair.lengthExtraBits()
	dSym := pair.distSymbol()
	dBits := pair.distExtraBits()
	return stats.llSymbols[lSym] + float64(lBits) + stats.dSymbols[dSym] + float64(dBits)
}

var dSymbolTable [30]uint16 = [30]uint16{
	1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513,
	769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577,
}

// Finds the minimum possible cost this cost model can return for valid length and
// distance symbols.
func (costModel costModelFun) minCost(costContext interface{}) float64 {
	var minCost float64

	// Table of distances that have a different distance symbol in the deflate
	// specification. Each value is the first distance that has a new symbol. Only
	// different symbols affect the cost model so only these need to be checked.
	// See RFC 1951 section 3.2.5. Compressed blocks (length and distance codes).

	// bestPair has lowest cost in the cost model
	var bestPair, pair lz77Pair
	pair.dist = 1
	minCost = math.Inf(1)
	for pair.litLen = uint16(3); pair.litLen < 259; pair.litLen++ {
		c := costModel(pair, costContext)
		if c < minCost {
			bestPair.litLen = pair.litLen
			minCost = c
		}
	}

	// TODO: try using bestPair.litlen instead of 3
	pair.litLen = 3
	minCost = math.Inf(1)
	for i := 0; i < 30; i++ {
		pair.dist = dSymbolTable[i]
		c := costModel(pair, costContext)
		if c < minCost {
			bestPair.dist = pair.dist
			minCost = c
		}
	}

	return costModel(bestPair, costContext)
}

// Performs the forward pass for "squeeze". Gets the most optimal length to reach
// every byte from a previous byte, using cost calculations.
// s: the BlockState
// inStart: where to start
// inEnd: where to stop (not inclusive)
// costModel: function to calculate the cost of some lit/len/dist pair.
// costContext: abstract context for the costmodel function
// lengths: output slice of size (inEnd - instart) which will receive the best length to reach this byte from a previous byte.
// returns the cost that was, according to the costmodel, needed to get to the end.
func (s *BlockState) bestLengths(inStart, inEnd int, costModel costModelFun, costContext interface{}) (lengths []uint16) {
	// Best cost to get here so far.
	if inStart == inEnd {
		return nil
	}

	blockSize := inEnd - inStart
	lengths = make([]uint16, blockSize+1)
	costs := make([]float64, blockSize+1)

	var windowStart int
	if inStart > WINDOW_SIZE {
		windowStart = inStart - WINDOW_SIZE
	}
	h := newHash(s.block[windowStart], s.block[windowStart+1])
	for i := windowStart; i < inStart; i++ {
		h.update(s.block, i, inEnd)
	}

	minCost := costModel.minCost(costContext)
	infinity := math.Inf(1)
	for i := 1; i <= blockSize; i++ {
		costs[i] = infinity
	}

	sublen := make([]uint16, 259)
	for i := inStart; i < inEnd; i++ {
		j := i - inStart // Index in the costs slice and lengths.
		h.update(s.block, i, inEnd)
		cost := costs[j]

		if SHORTCUT_LONG_REPETITIONS {
			// If we're in a long repetition of the same character and have
			// more than MAX_MATCH characters before and after our position.
			if h.same[i&WINDOW_MASK] > MAX_MATCH*2 &&
				i > inStart+MAX_MATCH+1 &&
				i+MAX_MATCH*2+1 < inEnd &&
				h.same[(i-MAX_MATCH)&WINDOW_MASK] > MAX_MATCH {
				symbolCost := costModel(lz77Pair{MAX_MATCH, 1}, costContext)
				// Set the length to reach each one to MAX_MATCH, and the cost
				// to the cost corresponding to that length. Doing this, we
				// skip MAX_MATCH values to avoid calling findLongestMatch.
				for k := 0; k < MAX_MATCH; k++ {
					costs[j+MAX_MATCH] = cost + symbolCost
					lengths[j+MAX_MATCH] = MAX_MATCH
					i++
					j++
					h.update(s.block, i, inEnd)
					cost = costs[j]
				}
			}
		}

		pair := s.findLongestMatch(&h, s.block, i, inEnd, MAX_MATCH, sublen)
		leng := pair.litLen

		// Literal.
		if i+1 <= inEnd {
			newCost := cost + costModel(lz77Pair{uint16(s.block[i]), 0}, costContext)
			if !(newCost >= 0) {
				panic("new cost is not positive")
			}
			if newCost < costs[j+1] {
				costs[j+1] = newCost
				lengths[j+1] = 1
			}
		}
		// Lengths.
		for k := uint16(3); k <= leng && i+int(k) <= inEnd; k++ {
			// Calling the cost model is expensive, avoid this if we are
			// already at the minimum possible cost that it can return.
			nextCost := costs[j+int(k)]
			if nextCost <= minCost+cost {
				continue
			}

			newCost := cost + costModel(lz77Pair{k, sublen[k]}, costContext)
			if !(newCost >= 0) {
				panic("new cost is not positive")
			}
			if newCost < nextCost {
				if k > MAX_MATCH {
					panic("k is larger than MAX_MATCH")
				}
				costs[j+int(k)] = newCost
				lengths[j+int(k)] = k
			}
		}
	}

	cost := costs[blockSize]
	if !(cost >= 0) {
		panic("cost is not positive")
	}
	if math.IsNaN(cost) {
		panic("cost is NaN")
	}
	if math.IsInf(cost, 0) {
		panic("cost is infinite")
	}
	return lengths
}

// Calculates the optimal path of lz77 lengths to use, from the calculated
// lengths. The lengths must contain the optimal length to reach that
// byte. The path will be filled with the lengths to use, so its data size will be
// the amount of lz77 symbols.
func traceBackwards(size int, lengths []uint16) []uint16 {
	if size == 0 {
		return nil
	}

	var path []uint16
	index := size
	for {
		path = append(path, lengths[index])
		if int(lengths[index]) > index {
			panic("length is greater than index")
		}
		if lengths[index] > MAX_MATCH {
			panic("length is greater than MAX_MATCH")
		}
		if lengths[index] == 0 {
			panic("length is zero")
		}
		index -= int(lengths[index])
		if index == 0 {
			break
		}
	}

	// Mirror result.
	pathSize := len(path)
	for index = 0; index < pathSize/2; index++ {
		path[index], path[pathSize-index-1] = path[pathSize-index-1], path[index]
	}

	return path
}

func (s *BlockState) followPath(inStart, inEnd int,
	path []uint16) LZ77Store {
	var store LZ77Store
	if inStart == inEnd {
		return store
	}

	var windowStart int
	if inStart > WINDOW_SIZE {
		windowStart = inStart - WINDOW_SIZE
	}
	h := newHash(s.block[windowStart], s.block[windowStart+1])
	for i := windowStart; i < inStart; i++ {
		h.update(s.block, i, inEnd)
	}

	pos := inStart
	for _, length := range path {
		if pos >= inEnd {
			panic("position overrun")
		}

		h.update(s.block, pos, inEnd)

		// Add to output.
		if length >= MIN_MATCH {
			// Get the distance by recalculating longest match. The
			// found length should match the length from the path.
			pair := s.findLongestMatch(&h, s.block, pos, inEnd, length, nil)
			if pair.litLen != length && length > 2 && pair.litLen > 2 {
				panic("dummy length is invalid")
			}
			pair.Verify(s.block, pos)
			store = append(store, pair)
		} else {
			length = 1
			store = append(store, lz77Pair{uint16(s.block[pos]), 0})
		}

		if pos+int(length) > inEnd {
			panic("position overrun")
		}
		for j := 1; j < int(length); j++ {
			h.update(s.block, pos+j, inEnd)
		}

		pos += int(length)
	}
	return store
}

// Calculates the entropy of the statistics
func (stats *symbolStats) calculate() {
	stats.llSymbols = CalculateEntropy(stats.litLens)
	stats.dSymbols = CalculateEntropy(stats.dists)
}

// Appends the symbol statistics from the store.
func (store LZ77Store) statistics() (stats symbolStats) {
	stats.litLens = make([]float64, 288)
	stats.dists = make([]float64, 32)
	storeSize := len(store)
	for i := 0; i < storeSize; i++ {
		if store[i].dist == 0 {
			stats.litLens[store[i].litLen]++
		} else {
			stats.litLens[store[i].lengthSymbol()]++
			stats.dists[store[i].distSymbol()]++
		}
	}
	stats.litLens[256] = 1 // End symbol.

	stats.calculate()
	return stats
}

// Does a single run for LZ77Optimal. For good compression, repeated runs
// with updated statistics should be performed.
//
// s: the block state
// inStart: where to start
// inEnd: where to stop (not inclusive)
// lengths: slice of size (inEnd - inStart) used to store lengths
// costModel: function to use as the cost model for this squeeze run
// costContext: abstract context for the costmodel function
// store: place to output the LZ77 data
// returns the cost that was, according to the costmodel, needed to get to the end.
// This is not the actual cost.
func (s *BlockState) lz77OptimalRun(inStart, inEnd int, costModel costModelFun, costContext interface{}) (store LZ77Store) {
	lengths := s.bestLengths(inStart, inEnd, costModel, costContext)
	path := traceBackwards(inEnd-inStart, lengths)
	store = s.followPath(inStart, inEnd, path)
	return store
}

// Calculates lit/len and dist pairs for given data.
// If instart is larger than 0, it uses values before instart as starting
// dictionary.
func (s *BlockState) LZ77Optimal(inStart, inEnd int) LZ77Store {
	// Dist to get to here with smallest cost.
	bestCost := uint64(math.MaxUint64)
	var lastCost uint64
	// Try randomizing the costs a bit once the size stabilizes.
	var randomize bool

	ranState := newRanState()

	// Do regular deflate, then loop multiple shortest path
	// runs, each time using the statistics of the previous run.

	// Initial run.
	bestStore := s.LZ77Greedy(inStart, inEnd)
	bestStats := bestStore.statistics()
	lastStats := bestStats

	// Repeat statistics with each time the cost model
	// from the previous stat run.
	for i := 0; i < s.options.NumIterations; i++ {
		store := s.lz77OptimalRun(inStart, inEnd, costStat, lastStats)
		cost := store.CalculateBlockSize(2)
		if s.options.VerboseMore || (s.options.Verbose && cost < bestCost) {
			fmt.Fprintf(os.Stderr, "Iteration %d: %d bit\n", i, int(cost))
		}
		stats := store.statistics()
		if cost < bestCost {
			// Copy to the output store.
			bestStore = store
			bestStats = stats
			bestCost = cost
		}
		if i > 5 && cost == lastCost {
			lastStats = bestStats
			lastStats.randomizeFreqs(ranState)
			lastStats.calculate()
			randomize = true
		} else {
			if randomize {
				// This makes it converge slower but better. Do it only once the
				// randomness kicks in so that if the user does few iterations, it gives a
				// better result sooner.
				stats = addWeightedFreqs(stats, 1.0, lastStats, 0.5)
				stats.calculate()
			}
			lastStats = stats
		}
		lastCost = cost
	}
	return bestStore
}

// Does the same as LZ77Optimal, but optimized for the fixed tree of the
// deflate standard.
// The fixed tree never gives the best compression. But this gives the best
// possible LZ77 encoding possible with the fixed tree.
// This does not create or output any fixed tree, only LZ77 data optimized for
// using with a fixed tree.
// If inStart is larger than 0, it uses values before inStart as starting
// dictionary.
func (s *BlockState) LZ77OptimalFixed(inStart, inEnd int) LZ77Store {
	// Dist to get to here with smallest cost.
	// Shortest path for fixed tree This one should give the shortest possible
	// result for fixed tree, no repeated runs are needed since the tree is known.
	return s.lz77OptimalRun(inStart, inEnd, costFixed, nil)
}
