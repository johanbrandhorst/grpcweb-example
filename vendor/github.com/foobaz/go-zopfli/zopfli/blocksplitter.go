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

// The "f" for the findMinimum function below.
// i: the current parameter of f(i)
// context: for your implementation
type findMinimumFun func(i int, context interface{}) uint64

// Finds minimum of function f(i) where i is of type int, f(i) is of type
// float64, i is in range start-end (excluding end).
func findMinimum(f findMinimumFun, context interface{}, start, end int) int {
	if end-start < 1024 {
		best := uint64(math.MaxUint64)
		result := start
		for i := start; i < end; i++ {
			v := f(i, context)
			if v < best {
				best = v
				result = i
			}
		}
		return result
	}

	// Try to find minimum faster by recursively checking multiple points.
	const NUM = 9 // Good value: 9.
	var p [NUM]int
	var vp [NUM]uint64
	lastBest := uint64(math.MaxUint64)
	pos := start

	for end-start > NUM {
		for i := 0; i < NUM; i++ {
			p[i] = start + (i+1)*((end-start)/(NUM+1))
			vp[i] = f(p[i], context)
		}
		var bestIndex int
		best := vp[0]
		for i := 1; i < NUM; i++ {
			if vp[i] < best {
				best = vp[i]
				bestIndex = i
			}
		}
		if best > lastBest {
			break
		}

		if bestIndex > 0 {
			start = p[bestIndex-1]
		}
		if bestIndex < NUM-1 {
			end = p[bestIndex+1]
		}

		pos = p[bestIndex]
		lastBest = best
	}
	return pos
}

// Returns estimated cost of a block in bits.	It includes the size to encode the
// tree and the size to encode all literal, length and distance symbols and their
// extra bits.
//
// litLens: lz77 lit/lengths
// dists: ll77 distances
func (store LZ77Store) estimateCost() uint64 {
	return store.CalculateBlockSize(2)
}

type splitCostContext struct {
	store      LZ77Store
	start, end int
}

// Gets the cost which is the sum of the cost of the left and the right section
// of the data.
// type: findMinimumFun
func splitCost(i int, context interface{}) uint64 {
	c := context.(*splitCostContext)
	a := c.store[c.start:i]
	b := c.store[i:c.end]
	return a.estimateCost() + b.estimateCost()
}

func addSorted(splitPoints []int, value int) []int {
	oldSize := len(splitPoints)
	splitPoints = append(splitPoints, value)
	for i := 0; i < oldSize; i++ {
		if splitPoints[i] > value {
			copy(splitPoints[i+1:], splitPoints[i:])
			splitPoints[i] = value
			break
		}
	}
	return splitPoints
}

// Prints the block split points as decimal and hex values in the terminal.
func (store LZ77Store) printBlockSplitPoints(lz77SplitPoints []int) {
	llSize := len(store)
	nLZ77Points := len(lz77SplitPoints)
	splitPoints := make([]int, 0, nLZ77Points)
	// The input is given as lz77 indices, but we want to see the
	// uncompressed index values.
	if nLZ77Points > 0 {
		var pos int
		for i := 0; i < llSize; i++ {
			var length int
			if store[i].dist == 0 {
				length = 1
			} else {
				length = int(store[i].litLen)
			}
			if lz77SplitPoints[len(splitPoints)] == i {
				splitPoints = append(splitPoints, pos)
				if len(splitPoints) >= nLZ77Points {
					break
				}
			}
			pos += length
		}
	}
	if len(splitPoints) != nLZ77Points {
		panic("number of points do not match")
	}

	fmt.Fprintf(os.Stderr, "block split points: ")
	for _, point := range splitPoints {
		fmt.Fprintf(os.Stderr, "%d ", point)
	}
	fmt.Fprintf(os.Stderr, "(hex:")
	for _, point := range splitPoints {
		fmt.Fprintf(os.Stderr, " %x", point)
	}
	fmt.Fprintf(os.Stderr, ")\n")
}

// Finds next block to try to split, the largest of the available ones.
// The largest is chosen to make sure that if only a limited amount of blocks is
// requested, their sizes are spread evenly.
// llSize: the size of the LL77 data, which is the size of the done array here.
// done: array indicating which blocks starting at that position are no longer
// splittable (splitting them increases rather than decreases cost).
// splitPoints: the splitpoints found so far.
// nPoints: the amount of splitpoints found so far.
// lStart: output variable, giving start of block.
// lEnd: output variable, giving end of block.
// returns 1 if a block was found, 0 if no block found (all are done).
func findLargestSplittableBlock(llSize int, done []bool, splitPoints []int) (lStart, lEnd int, found bool) {
	var longest int
	nPoints := len(splitPoints)
	for i := 0; i <= nPoints; i++ {
		var start, end int
		if i != 0 {
			start = splitPoints[i-1]
		}
		if i == nPoints {
			end = llSize - 1
		} else {
			end = splitPoints[i]
		}
		if !done[start] && end > longest+start {
			lStart = start
			lEnd = end
			found = true
			longest = end - start
		}
	}
	return lStart, lEnd, found
}

func (store LZ77Store) blockSplitLZ77(options *Options, maxBlocks int) []int {
	llSize := len(store)
	if llSize < 10 {
		// This code fails on tiny files.
		return nil
	}

	done := make([]bool, llSize)

	var splitPoints []int
	var lStart int
	lEnd := llSize
	for {
		if maxBlocks > 0 && len(splitPoints)+1 >= maxBlocks {
			break
		}

		var c splitCostContext
		c.store = store
		c.start = lStart
		c.end = lEnd
		if lStart >= lEnd {
			panic("overrun")
		}
		llPos := findMinimum(splitCost, &c, lStart+1, lEnd)

		if llPos <= lStart {
			panic("underrun")
		}
		if llPos >= lEnd {
			panic("overrun")
		}

		a := store[lStart:llPos]
		b := store[llPos:lEnd]
		splitCost := a.estimateCost() + b.estimateCost()
		both := store[lStart:lEnd]
		origCost := both.estimateCost()

		if splitCost > origCost || llPos == lStart+1 || llPos == lEnd {
			done[lStart] = true
		} else {
			splitPoints = addSorted(splitPoints, llPos)
		}

		var found bool
		lStart, lEnd, found = findLargestSplittableBlock(llSize, done, splitPoints)
		if !found {
			// No further split will probably reduce compression.
			break
		}

		if lEnd < lStart+10 {
			break
		}
	}

	if options.Verbose {
		store.printBlockSplitPoints(splitPoints)
	}
	return splitPoints
}

// Does blocksplitting on uncompressed data.
// The output splitpoints are indices in the uncompressed bytes.
//
// options: general program options.
// in: uncompressed input data
// inStart: where to start splitting
// inEnd: where to end splitting (not inclusive)
// maxBlocks: maximum amount of blocks to split into, or 0 for no limit
// splitPoints: dynamic array to put the resulting split point coordinates into.
//   The coordinates are indices in the input array.
func blockSplit(options *Options, in []byte, inStart, inEnd, maxBlocks int) []int {
	s := NewBlockState(options, in, inStart, inEnd)

	// Unintuitively, using a simple LZ77 method here instead of LZ77Optimal
	// results in better blocks.
	store := s.LZ77Greedy(inStart, inEnd)
	lz77SplitPoints := store.blockSplitLZ77(options, maxBlocks)

	// Convert LZ77 positions to positions in the uncompressed input.
	var splitPoints []int
	pos := inStart
	if len(lz77SplitPoints) > 0 {
		storeSize := len(store)
		for i := 0; i < storeSize; i++ {
			var length int
			if store[i].dist == 0 {
				length = 1
			} else {
				length = int(store[i].litLen)
			}
			if lz77SplitPoints[len(splitPoints)] == i {
				splitPoints = append(splitPoints, pos)
				if len(splitPoints) == len(lz77SplitPoints) {
					break
				}
			}
			pos += length
		}
	}
	if len(splitPoints) != len(lz77SplitPoints) {
		panic("number of points do not match")
	}
	return splitPoints
}

// Divides the input into equal blocks, does not even take LZ77 lengths into
// account.
func blockSplitSimple(inStart, inEnd, blockSize int) (splitPoints []int) {
	i := inStart
	for i < inEnd {
		splitPoints = append(splitPoints, i)
		i += blockSize
	}
	return splitPoints
}
