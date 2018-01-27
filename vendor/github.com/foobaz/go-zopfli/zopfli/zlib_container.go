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
	"hash/adler32"
	"io"
	"os"
)

func ZlibCompress(options *Options, in []byte, out io.Writer) error {
	var counter countingWriter
	if options.Verbose {
		counter = newCountingWriter(out)
		out = &counter
	}

	const cmf = 120 /* CM 8, CINFO 7. See zlib spec.*/
	const flevel = 0
	const fdict = 0
	var cmfflg uint16 = 256*cmf + fdict*32 + flevel*64
	fcheck := 31 - cmfflg%31
	cmfflg += fcheck
	flagBytes := []byte{
		byte(cmfflg >> 8),
		byte(cmfflg),
	}
	_, flagErr := out.Write(flagBytes)
	if flagErr != nil {
		return flagErr
	}

	z := NewDeflator(out, options)
	writeErr := z.Deflate(true, in)
	if writeErr != nil {
		return writeErr
	}

	checksum := adler32.New()
	checksum.Write(in)
	final := checksum.Sum32()
	checksumBytes := []byte{
		byte(final >> 24),
		byte(final >> 16),
		byte(final >> 8),
		byte(final),
	}
	_, checksumErr := out.Write(checksumBytes)
	if checksumErr != nil {
		return checksumErr
	}

	if options.Verbose {
		inSize := len(in)
		outSize := counter.written
		fmt.Fprintf(os.Stderr,
			"Original Size: %d, Zlib: %d, Compression: %f%% Removed\n",
			inSize, outSize,
			100*float64(inSize-outSize)/float64(inSize))
	}
	return nil
}
