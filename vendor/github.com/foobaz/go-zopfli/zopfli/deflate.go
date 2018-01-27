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
	"bufio"
	"fmt"
	"io"
	"os"
)

type lz77Lengths struct {
	litLen, dist []uint
}

type lz77Symbols struct {
	litLen, dist []uint
}

type Deflator struct {
	// out: pointer to the dynamic output array to which the result
	//   is appended. Must be freed after use.
	out *bufio.Writer

	// bp: number of bits written. This is because deflate appends
	//   blocks as bit-based data, rather than on byte boundaries.
	bp uint

	next    byte
	options *Options
}

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func NewDeflator(wr io.Writer, options *Options) Deflator {
	out := bufio.NewWriter(wr)
	return Deflator{out, 0, 0, options}
}

func (z *Deflator) writeBit(bit byte) {
	withinByte := z.bp & 7
	z.next |= (bit << withinByte)
	if withinByte == 7 {
		err := z.out.WriteByte(z.next)
		if err != nil {
			panic(err)
		}
		z.next = 0
	}
	z.bp++
}

func (z *Deflator) writeBits(symbol, length uint) {
	// TODO(lode): make more efficient (add more bits at once).
	for i := uint(0); i < length; i++ {
		bit := byte((symbol >> i) & 1)
		z.writeBit(bit)
	}
}

// Adds bits, like AddBits, but the order is inverted. The deflate specification
// uses both orders in one standard.
func (z *Deflator) writeHuffmanBits(symbol, length uint) {
	// TODO(lode): make more efficient (add more bits at once).
	for i := uint(0); i < length; i++ {
		bit := byte((symbol >> (length - i - 1)) & 1)
		z.writeBit(bit)
	}
}

func (z *Deflator) flush() {
	withinByte := z.bp & 7
	if withinByte > 0 {
		z.writeByte(z.next)
		z.next = 0
		z.bp += 8 - withinByte
	}
	err := z.out.Flush()
	if err != nil {
		panic(err)
	}
}

func (z *Deflator) writeByte(c byte) {
	err := z.out.WriteByte(c)
	if err != nil {
		panic(err)
	}
	z.bp += 8
}

func (z *Deflator) write(p []byte) {
	_, err := z.out.Write(p)
	if err != nil {
		panic(err)
	}
	z.bp += uint(len(p)) * 8
}

// Ensures there are at least 2 distance codes to support buggy decoders.
// Zlib 1.2.1 and below have a bug where it fails if there isn't at least 1
// distance code (with length > 0), even though it's valid according to the
// deflate spec to have 0 distance codes. On top of that, some mobile phones
// require at least two distance codes. To support these decoders too (but
// potentially at the cost of a few bytes), add dummy code lengths of 1.
// References to this bug can be found in the changelog of
// Zlib 1.2.2 and here: http://www.jonof.id.au/forum/index.php?topic=515.0.
//
// dLengths: the 32 lengths of the distance codes.
func (lengths lz77Lengths) patchDistanceCodesForBuggyDecoders() {
	var numDistCodes uint // Amount of non-zero distance codes
	dLengths := lengths.dist
	for i := 0; i < 30; /* Ignore the two unused codes from the spec */ i++ {
		if dLengths[i] > 0 {
			numDistCodes++
		}
		if numDistCodes >= 2 {
			// Two or more codes is fine.
			return
		}
	}

	if numDistCodes == 0 {
		dLengths[0] = 1
		dLengths[1] = 1
	} else if numDistCodes == 1 {
		var i int
		if dLengths[0] != 0 {
			i = 1
		}
		dLengths[i] = 1
	}
}

// The order in which code length code lengths are encoded as per deflate.
var clclOrder [19]uint8 = [19]uint8{
	16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15,
}

func (z *Deflator) writeDynamicTree(lengths lz77Lengths) {
	hLit := uint(29)  // 286 - 257
	hDist := uint(29) // 32 - 1, but gzip does not like hDist > 29.*/
	// Trim zeros.
	for hLit > 0 && lengths.litLen[257+hLit-1] == 0 {
		hLit--
	}
	for hDist > 0 && lengths.dist[1+hDist-1] == 0 {
		hDist--
	}

	// Size of lldLengths.
	lldTotal := hLit + 257 + hDist + 1

	// All litLen and dist lengths with ending
	// zeros trimmed together in one array.
	lldLengths := make([]uint, lldTotal)

	for i := uint(0); i < lldTotal; i++ {
		if i < 257+hLit {
			lldLengths[i] = lengths.litLen[i]
		} else {
			lldLengths[i] = lengths.dist[i-257-hLit]
		}
		if lldLengths[i] >= 16 {
			panic("length too large")
		}
	}

	// Runlength encoded version of lengths of litLen and dist trees.
	var rle []uint
	// Extra bits for rle values 16, 17 and 18.
	var rleBits []uint
	for i := uint(0); i < lldTotal; i++ {
		var count uint
		for j := i; j < lldTotal && lldLengths[i] == lldLengths[j]; j++ {
			count++
		}
		if count >= 4 || (count >= 3 && lldLengths[i] == 0) {
			if lldLengths[i] == 0 {
				if count > 10 {
					if count > 138 {
						count = 138
					}
					rle = append(rle, 18)
					rleBits = append(rleBits, count-11)
				} else {
					rle = append(rle, 17)
					rleBits = append(rleBits, count-3)
				}
			} else {
				rle = append(rle, lldLengths[i])
				rleBits = append(rleBits, 0)
				repeat := count - 1 // Since the first one is hardcoded.
				for repeat >= 6 {
					rle = append(rle, 16)
					rleBits = append(rleBits, 6-3)
					repeat -= 6
				}
				if repeat >= 3 {
					rle = append(rle, 16)
					rleBits = append(rleBits, repeat-3)
					repeat = 0
				}
				for repeat > 0 {
					rle = append(rle, lldLengths[i])
					rleBits = append(rleBits, 0)
					repeat--
				}
			}

			i += count - 1
		} else {
			rle = append(rle, lldLengths[i])
			rleBits = append(rleBits, 0)
		}
		if rle[len(rle)-1] > 18 {
			panic("last rle too large")
		}
	}

	rleSize := len(rle)
	var clCounts [19]uint
	for i := 0; i < rleSize; i++ {
		clCounts[rle[i]]++
	}

	// Code length code lengths.
	clcl := lengthLimitedCodeLengths(clCounts[:], 7)
	clSymbols := lengthsToSymbols(clcl, 7)

	// Trim zeros.
	hcLen := uint(15)
	for hcLen > 0 && clCounts[clclOrder[hcLen+4-1]] == 0 {
		hcLen--
	}

	z.writeBits(hLit, 5)
	z.writeBits(hDist, 5)
	z.writeBits(hcLen, 4)

	for i := uint(0); i < hcLen+4; i++ {
		z.writeBits(clcl[clclOrder[i]], 3)
	}

	for i := 0; i < rleSize; i++ {
		symbol := clSymbols[rle[i]]
		z.writeHuffmanBits(symbol, clcl[rle[i]])
		// Extra bits.
		if rle[i] == 16 {
			z.writeBits(rleBits[i], 2)
		} else if rle[i] == 17 {
			z.writeBits(rleBits[i], 3)
		} else if rle[i] == 18 {
			z.writeBits(rleBits[i], 7)
		}
	}
}

// Gives the exact size of the tree, in bits, as it will be encoded in DEFLATE.
func (lengths lz77Lengths) calculateTreeSize() uint {
	var w *nullWriter
	z := NewDeflator(w, nil)
	z.writeDynamicTree(lengths)
	return z.bp
}

func (lengths lz77Lengths) symbols(maxBits uint) (symbols lz77Symbols) {
	symbols.litLen = lengthsToSymbols(lengths.litLen, maxBits)
	symbols.dist = lengthsToSymbols(lengths.dist, maxBits)
	return symbols
}

// Adds all lit/len and dist codes from the lists as huffman symbols. Does not
// add end code 256. expectedDataSize is the uncompressed block size, used for
// assert, but you can set it to 0 to not do the assertion.
func (z *Deflator) writeLZ77Data(store LZ77Store,
	expectedDataSize int,
	symbols lz77Symbols, lengths lz77Lengths) {
	var testLength int
	lEnd := len(store)
	for i := 0; i < lEnd; i++ {
		pair := store[i]
		if pair.dist == 0 {
			if pair.litLen >= 256 {
				panic("litLen too large")
			}
			if lengths.litLen[pair.litLen] <= 0 {
				panic("length is zero")
			}
			z.writeHuffmanBits(symbols.litLen[pair.litLen], lengths.litLen[pair.litLen])
			testLength++
		} else {
			lls := pair.lengthSymbol()
			ds := pair.distSymbol()
			if pair.litLen < 3 || pair.litLen > 288 {
				panic("litLen out of range")
			}
			if lengths.litLen[lls] <= 0 {
				panic("length is zero")
			}
			if lengths.dist[ds] <= 0 {
				panic("length is zero")
			}
			z.writeHuffmanBits(symbols.litLen[lls], lengths.litLen[lls])
			z.writeBits(uint(pair.lengthExtraBitsValue()), uint(pair.lengthExtraBits()))
			z.writeHuffmanBits(symbols.dist[ds], lengths.dist[ds])
			z.writeBits(uint(pair.distExtraBitsValue()), uint(pair.distExtraBits()))
			testLength += int(pair.litLen)
		}
	}
	if expectedDataSize != 0 && testLength != expectedDataSize {
		panic("actual size did not match expected size")
	}
}

func getFixedTree() (lengths lz77Lengths) {
	lengths.litLen = make([]uint, 288)
	lengths.dist = make([]uint, 32)
	for i := 0; i < 144; i++ {
		lengths.litLen[i] = 8
	}
	for i := 144; i < 256; i++ {
		lengths.litLen[i] = 9
	}
	for i := 256; i < 280; i++ {
		lengths.litLen[i] = 7
	}
	for i := 280; i < 288; i++ {
		lengths.litLen[i] = 8
	}
	for i := 0; i < 32; i++ {
		lengths.dist[i] = 5
	}
	return lengths
}

// Calculates size of the part after the header and tree of an LZ77 block, in bits.
func (store LZ77Store) calculateBlockSymbolSize(lengths lz77Lengths) uint64 {
	var result uint64
	lEnd := len(store)
	for i := 0; i < lEnd; i++ {
		if store[i].dist == 0 {
			result += uint64(lengths.litLen[store[i].litLen])
		} else {
			result += uint64(lengths.litLen[store[i].lengthSymbol()])
			result += uint64(lengths.dist[store[i].distSymbol()])
			result += uint64(store[i].lengthExtraBits())
			result += uint64(store[i].distExtraBits())
		}
	}
	result += uint64(lengths.litLen[256]) // end symbol
	return result
}

// Calculates block size in bits.
// litLens: lz77 lit/lengths
// dists: ll77 distances
func (store LZ77Store) CalculateBlockSize(blockType byte) uint64 {
	if blockType != FIXED_BLOCK && blockType != DYNAMIC_BLOCK {
		panic("this is not for uncompressed blocks")
	}

	var lengths lz77Lengths
	result := uint64(3) // bFinal and blockType bits
	if blockType == FIXED_BLOCK {
		lengths = getFixedTree()
	} else {
		counts := store.lz77Counts()
		lengths.litLen = lengthLimitedCodeLengths(counts.litLen, 15)
		lengths.dist = lengthLimitedCodeLengths(counts.dist, 15)
		lengths.patchDistanceCodesForBuggyDecoders()
		result += uint64(lengths.calculateTreeSize())
	}

	result += store.calculateBlockSymbolSize(lengths)
	return result
}

// Adds a deflate block with the given LZ77 data to the output.
// z: the stream to write to
// blockType: the block type, must be 1 or 2
// final: whether to set the "final" bit on this block, must be the last block
// store: literal/length/distance array of the LZ77 data
// expectedDataSize: the uncompressed block size, used for panic, but you can
//   set it to 0 to not do the assertion.
func (z *Deflator) WriteLZ77Block(blockType byte, final bool, store LZ77Store, expectedDataSize int) {
	var finalByte byte
	if final {
		finalByte = 1
	}
	z.writeBit(finalByte)
	z.writeBit(blockType & 1)
	z.writeBit((blockType & 2) >> 1)

	var lengths lz77Lengths
	if blockType == FIXED_BLOCK {
		// Fixed block.
		lengths = getFixedTree()
	} else {
		// Dynamic block.
		if blockType != DYNAMIC_BLOCK {
			panic("illegal block type")
		}
		counts := store.lz77Counts()
		lengths.litLen = lengthLimitedCodeLengths(counts.litLen, 15)
		lengths.dist = lengthLimitedCodeLengths(counts.dist, 15)
		lengths.patchDistanceCodesForBuggyDecoders()
		detectTreeSize := z.bp
		z.writeDynamicTree(lengths)
		if z.options.Verbose {
			fmt.Fprintf(os.Stderr, "treesize: %d bits\n", z.bp-detectTreeSize)
		}

		// Assert that for every present symbol, the code length is non-zero.
		// TODO(lode): remove this in release version.
		for i := 0; i < 288; i++ {
			if counts.litLen[i] != 0 && lengths.litLen[i] <= 0 {
				panic("length is zero")
			}
		}
		for i := 0; i < 32; i++ {
			if counts.dist[i] != 0 && lengths.dist[i] <= 0 {
				panic("length is zero")
			}
		}
	}

	symbols := lengths.symbols(15)

	detectBlockSize := z.bp
	z.writeLZ77Data(store, expectedDataSize, symbols, lengths)
	// End symbol.
	z.writeHuffmanBits(symbols.litLen[256], lengths.litLen[256])
	if final {
		// write last byte
		z.flush()
	}

	if z.options.Verbose {
		var uncompressedSize uint
		lEnd := len(store)
		for i := 0; i < lEnd; i++ {
			if store[i].dist == 0 {
				uncompressedSize += 1
			} else {
				uncompressedSize += uint(store[i].litLen)
			}
		}
		compressedSize := z.bp - detectBlockSize
		var places int
		if compressedSize&1 != 0 {
			places = 3
		} else if compressedSize&2 != 0 {
			places = 2
		} else if compressedSize&4 != 0 {
			places = 1
		}
		fmt.Fprintf(
			os.Stderr,
			"compressed block size: %.*f (%dkB) (unc: %d (%dkB)\n",
			places,
			float64(compressedSize)/8,
			(compressedSize+4000)/8000,
			uncompressedSize,
			(uncompressedSize+500)/1000,
		)
	}
}

func (z *Deflator) deflateDynamicBlock(final bool, in []byte, inStart, inEnd int) {
	s := NewBlockState(z.options, in, inStart, inEnd)
	store := s.LZ77Optimal(inStart, inEnd)

	// For small block, encoding with fixed tree can be smaller. For large block,
	// don't bother doing this expensive test, dynamic tree will be better.
	blockType := byte(DYNAMIC_BLOCK)
	if len(store) < 1000 {
		fixedStore := s.LZ77OptimalFixed(inStart, inEnd)
		dynCost := store.CalculateBlockSize(2)
		fixedCost := fixedStore.CalculateBlockSize(1)
		if fixedCost < dynCost {
			blockType = FIXED_BLOCK
			store = fixedStore
		}
	}

	blockSize := inEnd - inStart
	z.WriteLZ77Block(blockType, final, store, blockSize)
}

func (z *Deflator) deflateFixedBlock(final bool, in []byte, inStart, inEnd int) {
	blockSize := inEnd - inStart

	s := NewBlockState(z.options, in, inStart, inEnd)
	store := s.LZ77OptimalFixed(inStart, inEnd)
	z.WriteLZ77Block(FIXED_BLOCK, final, store, blockSize)
}

func (z *Deflator) deflateNonCompressedBlock(final bool, in []byte) {
	blockSize := len(in)
	if blockSize >= 65536 {
		panic("Non compressed blocks are max this size.")
	}
	nLen := uint16(^blockSize)

	var finalByte byte
	if final {
		finalByte = 1
	}
	z.writeBit(finalByte)
	// blockType 00
	z.writeBit(0)
	z.writeBit(0)
	// Any bits of input up to the next byte boundary are ignored.
	z.flush()

	z.writeByte(byte(blockSize))
	z.writeByte(byte(blockSize / 256))
	z.writeByte(byte(nLen))
	z.writeByte(byte(nLen / 256))

	z.write(in)
}

func (z *Deflator) deflateBlock(final bool, in []byte, inStart, inEnd int) {
	switch z.options.BlockType {
	case UNCOMPRESSED_BLOCK:
		z.deflateNonCompressedBlock(final, in[inStart:inEnd])
	case FIXED_BLOCK:
		z.deflateFixedBlock(final, in, inStart, inEnd)
	case DYNAMIC_BLOCK:
		z.deflateDynamicBlock(final, in, inStart, inEnd)
	default:
		panic("illegal block type")
	}
}

// Does squeeze strategy where first block splitting is done, then each block is
// squeezed.
// Parameters: see description of the Deflate function.
func (z *Deflator) deflateSplittingFirst(final bool, in []byte, inStart, inEnd int) {
	var splitPoints []int
	switch z.options.BlockType {
	case UNCOMPRESSED_BLOCK:
		splitPoints = blockSplitSimple(inStart, inEnd, 65535)
	case FIXED_BLOCK:
		// If all blocks are fixed tree, splitting into separate blocks only
		// increases the total size. Leave splitPoints nil, this represents 1 block.
	case DYNAMIC_BLOCK:
		splitPoints = blockSplit(z.options, in, inStart, inEnd, z.options.BlockSplittingMax)
	}

	nPoints := len(splitPoints)
	for i := 0; i <= nPoints; i++ {
		var start, end int
		if i == 0 {
			start = inStart
		} else {
			start = splitPoints[i-1]
		}
		if i == nPoints {
			end = inEnd
		} else {
			end = splitPoints[i]
		}
		z.deflateBlock(i == nPoints && final, in, start, end)
	}
}

// Does squeeze strategy where first the best possible lz77 is done, and then based
// on that data, block splitting is done.
// Parameters: see description of the Deflate function.
func (z *Deflator) deflateSplittingLast(final bool, in []byte, inStart, inEnd int) {
	blockType := z.options.BlockType
	if blockType == UNCOMPRESSED_BLOCK {
		// This function only supports LZ77 compression. deflateSplittingFirst
		// supports the special case of noncompressed data. Punt it to that one.
		z.deflateSplittingFirst(final, in, inStart, inEnd)
		return
	}
	if blockType != FIXED_BLOCK && blockType != DYNAMIC_BLOCK {
		panic("illegal block type")
	}

	s := NewBlockState(z.options, in, inStart, inEnd)

	var store LZ77Store
	if blockType == DYNAMIC_BLOCK {
		store = s.LZ77Optimal(inStart, inEnd)
	} else {
		if blockType != FIXED_BLOCK {
			panic("illegal block type")
		}
		store = s.LZ77OptimalFixed(inStart, inEnd)
	}

	// If all blocks are fixed tree, splitting into separate blocks only
	// increases the total size. Leave nPoints at 0, this represents 1 block.
	var splitPoints []int
	if blockType != FIXED_BLOCK {
		splitPoints = store.blockSplitLZ77(z.options, z.options.BlockSplittingMax)
	}

	storeSize := len(store)
	nPoints := len(splitPoints)
	for i := 0; i <= nPoints; i++ {
		var start, end int
		if i > 0 {
			start = splitPoints[i-1]
		}
		if i >= nPoints {
			end = storeSize
		} else {
			end = splitPoints[i]
		}
		z.WriteLZ77Block(blockType, i == nPoints && final, store[start:end], 0)
	}
}

// Deflate a part, to allow Deflate() to use multiple master blocks if
// needed.
//
// Like Deflate, but allows to specify start and end byte with inStart and
// inEnd. Only that part is compressed, but earlier bytes are still used for the
// back window.
//
// It is possible to call this function multiple times in a row, shifting
// inStart and inEnd to next bytes of the data. If inStart is larger than 0, then
// previous bytes are used as the initial dictionary for LZ77.
// This function will usually output multiple deflate blocks. If final is 1, then
// the final bit will be set on the last block.
func (z *Deflator) DeflatePart(final bool, in []byte, inStart, inEnd int) (err error) {
	defer func() {
		problem := recover()
		if problem != nil {
			err = problem.(error)
		}
	}()

	if z.options.BlockSplitting {
		if z.options.BlockSplittingLast {
			z.deflateSplittingLast(final, in, inStart, inEnd)
		} else {
			z.deflateSplittingFirst(final, in, inStart, inEnd)
		}
	} else {
		z.deflateBlock(final, in, inStart, inEnd)
	}
	return err
}

// Compresses according to the deflate specification and append the compressed
// result to the output.
// This function will usually output multiple deflate blocks. If final is 1, then
// the final bit will be set on the last block.
//
// final: whether this is the last section of the input, sets the final bit to the
//	 last deflate block.
// in: the input bytes
func (z *Deflator) Deflate(final bool, in []byte) (err error) {
	defer func() {
		problem := recover()
		if problem != nil {
			err = problem.(error)
		}
	}()

	if MASTER_BLOCK_SIZE == 0 {
		err := z.DeflatePart(true, in, 0, len(in))
		if err != nil {
			return err
		}
	} else {
		var i int
		inSize := len(in)
		for i < inSize {
			var size int
			masterFinal := i+MASTER_BLOCK_SIZE >= inSize
			final2 := final && masterFinal
			if masterFinal {
				size = inSize - i
			} else {
				size = MASTER_BLOCK_SIZE
			}
			err := z.DeflatePart(final2, in, i, i+size)
			if err != nil {
				return err
			}
			i += size
		}
	}
	if z.options.Verbose {
		inSize := len(in)
		outSize := z.bp / 8
		fmt.Fprintf(
			os.Stderr,
			"Original Size: %d, Deflate: %d, Compression: %f%% Removed\n",
			inSize, outSize,
			100*(float64(inSize)-float64(outSize))/float64(inSize),
		)
	}
	return nil
}
