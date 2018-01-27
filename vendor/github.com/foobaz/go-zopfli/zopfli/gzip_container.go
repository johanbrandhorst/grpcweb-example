/*
Copyright 2013 Google Inc. All Rights Reserved.

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
	"hash/crc32"
	"io"
	"os"
)

// Compresses according to the gzip specification and writes the compressed
// result to the output.
//
// options: global program options
// out: writer to which the result is appended
func GzipCompress(options *Options, in []byte, out io.Writer) error {
	var counter countingWriter
	if options.Verbose {
		counter = newCountingWriter(out)
		out = &counter
	}

	header := []byte{
		// ID
		31,
		139,
		// CM
		8,
		// FLG
		0,
		// MTIME
		0,
		0,
		0,
		0,
		// XFL, 2 indicates best compression.
		2,
		// OS follows Unix conventions.
		3,
	}
	_, headerErr := out.Write(header)
	if headerErr != nil {
		return headerErr
	}

	z := NewDeflator(out, options)
	writeErr := z.Deflate(true, in)
	if writeErr != nil {
		return writeErr
	}

	checksum := crc32.NewIEEE()
	checksum.Write(in)
	crcValue := checksum.Sum32()
	inSize := len(in)
	footer := []byte{
		// CRC
		byte(crcValue),
		byte(crcValue >> 8),
		byte(crcValue >> 16),
		byte(crcValue >> 24),
		// ISIZE
		byte(inSize),
		byte(inSize >> 8),
		byte(inSize >> 16),
		byte(inSize >> 24),
	}
	_, footerErr := out.Write(footer)
	if footerErr != nil {
		return footerErr
	}

	if options.Verbose {
		inSize := len(in)
		outSize := counter.written
		fmt.Fprintf(os.Stderr,
			"Original Size: %d, Gzip: %d, Compression: %f%% Removed\n",
			inSize, outSize,
			100*float64(inSize-outSize)/float64(inSize))
	}
	return nil
}
