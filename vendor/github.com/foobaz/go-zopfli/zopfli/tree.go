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
	"math"
)

// Converts a series of Huffman tree bitLengths, to the bit values of the symbols.
func lengthsToSymbols(lengths []uint, maxBits uint) (symbols []uint) {
	n := len(lengths)
	blCount := make([]uint, maxBits+1)
	nextCode := make([]uint, maxBits+1)

	symbols = make([]uint, n)

	// 1) Count the number of codes for each code length.
	// Let blCount[N] be the number of codes of length N, N >= 1.
	for bits := uint(0); bits <= maxBits; bits++ {
		blCount[bits] = 0
	}
	for i := 0; i < n; i++ {
		if lengths[i] > maxBits {
			panic("length is too large")
		}
		blCount[lengths[i]]++
	}
	// 2) Find the numerical value of the smallest code for each code length.
	var code uint
	blCount[0] = 0
	for bits := uint(1); bits <= maxBits; bits++ {
		code = (code + blCount[bits-1]) << 1
		nextCode[bits] = code
	}
	// 3) Assign numerical values to all codes, using consecutive values for
	// all codes of the same length with the base values determined at step 2.
	for i := 0; i < n; i++ {
		len := lengths[i]
		if len != 0 {
			symbols[i] = nextCode[len]
			nextCode[len]++
		}
	}
	return symbols
}

// Calculates the entropy of each symbol, based on the counts of each symbol. The
// result is similar to the result of CalculateBitLengths, but with the
// actual theoritical bit lengths according to the entropy. Since the resulting
// values are fractional, they cannot be used to encode the tree specified by
// DEFLATE.
func CalculateEntropy(count []float64) (bitLengths []float64) {
	var sum, log2sum float64
	n := len(count)
	for i := 0; i < n; i++ {
		sum += count[i]
	}
	if sum == 0 {
		log2sum = math.Log2(float64(n))
	} else {
		log2sum = math.Log2(sum)
	}
	bitLengths = make([]float64, n)
	for i := 0; i < n; i++ {
		// When the count of the symbol is 0, but its cost is requested anyway, it
		// means the symbol will appear at least once anyway, so give it the cost as if
		// its count is 1.
		if count[i] == 0 {
			bitLengths[i] = log2sum
		} else {
			bitLengths[i] = math.Log2(sum / count[i])
		}
		if !(bitLengths[i] >= 0) {
			panic("bit length is not positive")
		}
	}
	return bitLengths
}
