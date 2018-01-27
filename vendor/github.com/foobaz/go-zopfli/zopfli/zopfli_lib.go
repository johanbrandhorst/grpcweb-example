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
	"io"
)

type countingWriter struct {
	w       io.Writer
	written int
}

func newCountingWriter(w io.Writer) countingWriter {
	return countingWriter{w, 0}
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	cw.written += len(p)
	return cw.w.Write(p)
}

func Compress(options *Options, outputType int, in []byte, out io.Writer) error {
	switch outputType {
	case FORMAT_GZIP:
		return GzipCompress(options, in, out)
	case FORMAT_ZLIB:
		return ZlibCompress(options, in, out)
	case FORMAT_DEFLATE:
		return DeflateCompress(options, in, out)
	}
	panic("Unknown output type")
}

func DeflateCompress(options *Options, in []byte, out io.Writer) error {
	z := NewDeflator(out, options)
	deflateErr := z.Deflate(true, in)
	if deflateErr != nil {
		return deflateErr
	}

	return nil
}
