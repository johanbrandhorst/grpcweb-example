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

const (
	// Minimum and maximum length that can be encoded in deflate.
	MAX_MATCH = 258
	MIN_MATCH = 3

	// The window size for deflate. Must be a power of two. This should be
	// 32768, the maximum possible by the deflate spec. Anything less hurts
	// compression more than speed.
	WINDOW_SIZE = 32768

	// The window mask used to wrap indices into the window. This is why the
	// window size must be a power of two.
	WINDOW_MASK = (WINDOW_SIZE - 1)

	// A block structure of huge, non-smart, blocks to divide the input into, to allow
	// operating on huge files without exceeding memory, such as the 1GB wiki9 corpus.
	// The whole compression algorithm, including the smarter block splitting, will
	// be executed independently on each huge block.
	// Dividing into huge blocks hurts compression, but not much relative to the size.
	// Set this to, for example, 20MB (20000000). Set it to 0 to disable master blocks.
	MASTER_BLOCK_SIZE = 20000000

	// For longest match cache. max 256. Uses huge amounts of memory but makes it
	// faster. Uses this many times three bytes per single byte of the input data.
	// This is so because longest match finding has to find the exact distance
	// that belongs to each length for the best lz77 strategy.
	// Good values: e.g. 5, 8.
	CACHE_LENGTH = 8

	// limit the max hash chain hits for this hash value. This has an effect only
	// on files where the hash value is the same very often. On these files, this
	// gives worse compression (the value should ideally be 32768, which is the
	// WINDOW_SIZE, while zlib uses 4096 even for best level), but makes it
	// faster on some specific files.
	// Good value: e.g. 8192.
	MAX_CHAIN_HITS = 8192

	// Whether to use the longest match cache for FindLongestMatch. This cache
	// consumes a lot of memory but speeds it up. No effect on compression size.
	LONGEST_MATCH_CACHE = true

	// Enable to remember amount of successive identical bytes in the hash chain for
	// finding longest match
	// required for HASH_SAME_HASH and SHORTCUT_LONG_REPETITIONS
	// This has no effect on the compression result, and enabling it increases speed.
	HASH_SAME = true

	// Switch to a faster hash based on the info from HASH_SAME once the
	// best length so far is long enough. This is way faster for files with lots of
	// identical bytes, on which the compressor is otherwise too slow. Regular files
	// are unaffected or maybe a tiny bit slower.
	// This has no effect on the compression result, only on speed.
	HASH_SAME_HASH = true

	// Enable this, to avoid slowness for files which are a repetition of the same
	// character more than a multiple of MAX_MATCH times. This should not affect
	// the compression result.
	SHORTCUT_LONG_REPETITIONS = true

	// Whether to use lazy matching in the greedy LZ77 implementation. This gives a
	// better result of LZ77Greedy, but the effect this has on the optimal LZ77
	// varies from file to file.
	LAZY_MATCHING = true
)

// Gets the amount of extra bits for the given dist, cfr. the DEFLATE spec.
func (pair lz77Pair) distExtraBits() uint16 {
	dist := pair.dist
	if dist < 5 {
		return 0
	} else if dist < 9 {
		return 1
	} else if dist < 17 {
		return 2
	} else if dist < 33 {
		return 3
	} else if dist < 65 {
		return 4
	} else if dist < 129 {
		return 5
	} else if dist < 257 {
		return 6
	} else if dist < 513 {
		return 7
	} else if dist < 1025 {
		return 8
	} else if dist < 2049 {
		return 9
	} else if dist < 4097 {
		return 10
	} else if dist < 8193 {
		return 11
	} else if dist < 16385 {
		return 12
	}
	return 13
}

// Gets value of the extra bits for the given dist, cfr. the DEFLATE spec.
func (pair lz77Pair) distExtraBitsValue() uint16 {
	dist := pair.dist
	switch {
	case dist < 5:
		return 0
	case dist < 9:
		return (dist - 5) & 1
	case dist < 17:
		return (dist - 9) & 3
	case dist < 33:
		return (dist - 17) & 7
	case dist < 65:
		return (dist - 33) & 15
	case dist < 129:
		return (dist - 65) & 31
	case dist < 257:
		return (dist - 129) & 63
	case dist < 513:
		return (dist - 257) & 127
	case dist < 1025:
		return (dist - 513) & 255
	case dist < 2049:
		return (dist - 1025) & 511
	case dist < 4097:
		return (dist - 2049) & 1023
	case dist < 8193:
		return (dist - 4097) & 2047
	case dist < 16385:
		return (dist - 8193) & 4095
	}
	return dist - 16385&8191
}

// Gets the symbol for the given dist, cfr. the DEFLATE spec.
func (pair lz77Pair) distSymbol() uint16 {
	dist := pair.dist
	if dist < 193 {
		if dist < 13 {
			// dist 0..13.
			if dist < 5 {
				return dist - 1
			} else if dist < 7 {
				return 4
			} else if dist < 9 {
				return 5
			}
			return 6
		} else {
			// dist 13..193.
			if dist < 17 {
				return 7
			} else if dist < 25 {
				return 8
			} else if dist < 33 {
				return 9
			} else if dist < 49 {
				return 10
			} else if dist < 65 {
				return 11
			} else if dist < 97 {
				return 12
			} else if dist < 129 {
				return 13
			}
			return 14
		}
	}
	if dist < 2049 {
		// dist 193..2049.
		if dist < 257 {
			return 15
		} else if dist < 385 {
			return 16
		} else if dist < 513 {
			return 17
		} else if dist < 769 {
			return 18
		} else if dist < 1025 {
			return 19
		} else if dist < 1537 {
			return 20
		}
		return 21
	}
	// dist 2049..32768.
	if dist < 3073 {
		return 22
	} else if dist < 4097 {
		return 23
	} else if dist < 6145 {
		return 24
	} else if dist < 8193 {
		return 25
	} else if dist < 12289 {
		return 26
	} else if dist < 16385 {
		return 27
	} else if dist < 24577 {
		return 28
	}
	return 29
}

var lengthExtraBitsTable [259]uint16 = [259]uint16{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 0,
}

// Gets the amount of extra bits for the given length, cfr. the DEFLATE spec.
func (pair lz77Pair) lengthExtraBits() uint16 {
	return lengthExtraBitsTable[pair.litLen]
}

var lengthExtraBitsValueTable [259]uint16 = [259]uint16{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 2, 3, 0,
	1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3, 0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5,
	6, 7, 0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
	13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2,
	3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28,
	29, 30, 31, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17,
	18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 0, 1, 2, 3, 4, 5, 6,
	7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
	27, 28, 29, 30, 31, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 0,
}

// Gets value of the extra bits for the given length, cfr. the DEFLATE spec.
func (pair lz77Pair) lengthExtraBitsValue() uint16 {
	return lengthExtraBitsValueTable[pair.litLen]
}

var lengthSymbolTable [259]uint16 = [259]uint16{
	0, 0, 0, 257, 258, 259, 260, 261, 262, 263, 264,
	265, 265, 266, 266, 267, 267, 268, 268,
	269, 269, 269, 269, 270, 270, 270, 270,
	271, 271, 271, 271, 272, 272, 272, 272,
	273, 273, 273, 273, 273, 273, 273, 273,
	274, 274, 274, 274, 274, 274, 274, 274,
	275, 275, 275, 275, 275, 275, 275, 275,
	276, 276, 276, 276, 276, 276, 276, 276,
	277, 277, 277, 277, 277, 277, 277, 277,
	277, 277, 277, 277, 277, 277, 277, 277,
	278, 278, 278, 278, 278, 278, 278, 278,
	278, 278, 278, 278, 278, 278, 278, 278,
	279, 279, 279, 279, 279, 279, 279, 279,
	279, 279, 279, 279, 279, 279, 279, 279,
	280, 280, 280, 280, 280, 280, 280, 280,
	280, 280, 280, 280, 280, 280, 280, 280,
	281, 281, 281, 281, 281, 281, 281, 281,
	281, 281, 281, 281, 281, 281, 281, 281,
	281, 281, 281, 281, 281, 281, 281, 281,
	281, 281, 281, 281, 281, 281, 281, 281,
	282, 282, 282, 282, 282, 282, 282, 282,
	282, 282, 282, 282, 282, 282, 282, 282,
	282, 282, 282, 282, 282, 282, 282, 282,
	282, 282, 282, 282, 282, 282, 282, 282,
	283, 283, 283, 283, 283, 283, 283, 283,
	283, 283, 283, 283, 283, 283, 283, 283,
	283, 283, 283, 283, 283, 283, 283, 283,
	283, 283, 283, 283, 283, 283, 283, 283,
	284, 284, 284, 284, 284, 284, 284, 284,
	284, 284, 284, 284, 284, 284, 284, 284,
	284, 284, 284, 284, 284, 284, 284, 284,
	284, 284, 284, 284, 284, 284, 284, 285,
}

// Gets the symbol for the given length, cfr. the DEFLATE spec.
// Returns the symbol in the range [257-285] (inclusive)
func (pair lz77Pair) lengthSymbol() uint16 {
	return lengthSymbolTable[pair.litLen]
}

func DefaultOptions() (options Options) {
	options.Verbose = false
	options.VerboseMore = false
	options.NumIterations = 15
	options.BlockSplitting = true
	options.BlockSplittingLast = false
	options.BlockSplittingMax = 15
	options.BlockType = DYNAMIC_BLOCK
	return options
}
