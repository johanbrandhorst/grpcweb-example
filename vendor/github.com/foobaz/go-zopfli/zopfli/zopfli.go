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

// Options used throughout the program.
type Options struct {
	// Whether to print output
	Verbose bool

	// Whether to print more detailed output
	VerboseMore bool

	// Maximum amount of times to rerun forward and backward pass to optimize
	// LZ77 compression cost. Good values: 10, 15 for small files, 5 for files
	// over several MB in size or it will be too slow.
	NumIterations int

	// If true, splits the data in multiple deflate blocks with optimal choice
	// for the block boundaries. Block splitting gives better compression. Default:
	// true.
	BlockSplitting bool

	// If true, chooses the optimal block split points only after doing the iterative
	// LZ77 compression. If false, chooses the block split points first, then does
	// iterative LZ77 on each individual block. Depending on the file, either first
	// or last gives the best compression. Default: false.
	BlockSplittingLast bool

	// Maximum amount of blocks to split into (0 for unlimited, but this can give
	// extreme results that hurt compression on some files). Default value: 15.
	BlockSplittingMax int

	// The deflate block type. Use 2 for best compression.
	//	 -0: non compressed blocks (00)
	//	 -1: blocks with fixed tree (01)
	//	 -2: blocks with dynamic tree (10)
	BlockType byte
}

// Output format
const (
	FORMAT_GZIP = iota
	FORMAT_ZLIB
	FORMAT_DEFLATE
)

// Block type
const (
	UNCOMPRESSED_BLOCK = iota
	FIXED_BLOCK
	DYNAMIC_BLOCK
)
