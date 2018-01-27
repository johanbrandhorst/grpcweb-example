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

/*
Bounded package merge algorithm, based on the paper
"A Fast and Space-Economical Algorithm for Length-Limited Coding
Jyrki Katajainen, Alistair Moffat, Andrew Turpin".
*/

package zopfli

import (
	"sort"
)

// Nodes forming chains. Also used to represent leaves.
type node struct {
	weight uint  // Total weight (symbol count) of this chain.
	tail   *node // Previous node(s) of this chain, or nil if none.
	count  int   // Leaf symbol index, or number of leaves before this chain.
}

type symbolLeaves []*node

// Initializes a chain node with the given values and marks it as in use.
func newNode(weight uint, count int, tail *node) *node {
	var node node
	node.weight = weight
	node.count = count
	node.tail = tail
	return &node
}

// Performs a Boundary Package-Merge step. Puts a new chain in the given list.
// The new chain is, depending on the weights, a leaf or a combination of two
// chains from the previous list.
// lists: The lists of chains.
// maxBits: Number of lists.
// leaves: The leaves, one per symbol.
// numSymbols: Number of leaves.
// index: The index of the list in which a new chain or leaf is required.
// final: Whether this is the last time this function is called. If it is then
// it is no more needed to recursively call self.
func boundaryPM(lists [][2]*node, leaves symbolLeaves, index int, final bool) {
	lastCount := lists[index][1].count // Count of last chain of list.

	numSymbols := len(leaves)
	if index == 0 && lastCount >= numSymbols {
		return
	}

	lists[index][0] = lists[index][1]

	if index == 0 {
		// New leaf node in list 0.
		lists[index][1] = newNode(leaves[lastCount].weight, lastCount+1, nil)
	} else {
		sum := lists[index-1][0].weight + lists[index-1][1].weight
		if lastCount < numSymbols && sum > leaves[lastCount].weight {
			// New leaf inserted in list, so count is incremented.
			lists[index][1] = newNode(leaves[lastCount].weight,
				lastCount+1, lists[index][1].tail)
		} else {
			lists[index][1] = newNode(sum, lastCount, lists[index-1][1])
			if !final {
				// Two lookahead chains of previous list used up, create new ones.
				boundaryPM(lists, leaves, index-1, false)
				boundaryPM(lists, leaves, index-1, false)
			}
		}
	}
}

// Initializes each list with as lookahead chains the two leaves with lowest
// weights.
func newLists(leaves symbolLeaves, maxBits int) (lists [][2]*node) {
	lists = make([][2]*node, maxBits)
	node0 := newNode(leaves[0].weight, 1, nil)
	node1 := newNode(leaves[1].weight, 2, nil)
	for i := 0; i < maxBits; i++ {
		lists[i][0] = node0
		lists[i][1] = node1
	}
	return lists
}

// Converts result of boundary package-merge to the bitLengths. The result in the
// last chain of the last list contains the amount of active leaves in each list.
// chain: Chain to extract the bit length from (last chain from last list).
func extractBitLengths(chain *node, leaves symbolLeaves, bitLengths []uint) {
	for node := chain; node != nil; node = node.tail {
		for i := 0; i < node.count; i++ {
			bitLengths[leaves[i].count]++
		}
	}
}

func (leaves *symbolLeaves) Len() int {
	return len(*leaves)
}

// Comparator for sorting the leaves. Has the function signature for qsort.
func (leaves *symbolLeaves) Less(i, j int) bool {
	return (*leaves)[i].weight < (*leaves)[j].weight
}

func (leaves *symbolLeaves) Swap(i, j int) {
	(*leaves)[j], (*leaves)[i] = (*leaves)[i], (*leaves)[j]
}

// Outputs minimum-redundancy length-limited code bitLengths for symbols with the
// given counts. The bitLengths are limited by maxBits.
//
// The output is tailored for DEFLATE: symbols that never occur, get a bit length
// of 0, and if only a single symbol occurs at least once, its bitlength will be 1,
// and not 0 as would theoretically be needed for a single symbol.
//
// frequencies: The amount of occurances of each symbol.
// n: The amount of symbols.
// maxBits: Maximum bit length, inclusive.
// bitLengths: Output, the bitlengths for the symbol prefix codes.
// return: 0 for OK, non-0 for error.
func lengthLimitedCodeLengths(frequencies []uint, maxBits int) []uint {
	n := len(frequencies)
	// One leaf per symbol. Only numSymbols leaves will be used.
	leaves := make(symbolLeaves, 0, n)
	// Count used symbols and place them in the leaves.
	for i := 0; i < n; i++ {
		if frequencies[i] > 0 {
			node := newNode(frequencies[i], i, nil)
			leaves = append(leaves, node)
		}
	}

	// Amount of symbols with frequency > 0.
	numSymbols := len(leaves)
	// Check special cases and error conditions.
	if (1 << uint(maxBits)) < numSymbols {
		// Error, too few maxBits to represent symbols.
		panic("couldn't calculate code lengths")
	}

	// Initialize all bitlengths at 0.
	bitLengths := make([]uint, n)
	if numSymbols == 0 {
		// No symbols at all. OK.
		return bitLengths
	}
	if numSymbols == 1 {
		// Only one symbol, give it bitLength 1, not 0. OK.
		bitLengths[leaves[0].count] = 1
		return bitLengths
	}

	// Sort the leaves from lightest to heaviest.
	sort.Sort(&leaves)

	// Array of lists of chains. Each list requires only two lookahead
	// chains at a time, so each list is a array of two node*'s.
	lists := newLists(leaves, maxBits)

	// In the last list, 2 * numSymbols - 2 active chains need to be created. Two
	// are already created in the initialization. Each boundaryPM run creates one.
	numBoundaryPMRuns := 2*numSymbols - 4
	for i := 0; i < numBoundaryPMRuns; i++ {
		final := i == numBoundaryPMRuns-1
		boundaryPM(lists, leaves, maxBits-1, final)
	}

	extractBitLengths(lists[maxBits-1][1], leaves, bitLengths)
	return bitLengths
}
