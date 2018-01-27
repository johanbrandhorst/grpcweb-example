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
	HASH_SHIFT = 5
	HASH_MASK  = 32767
)

type hash struct {
	head    []int    // Hash value to index of its most recent occurance.
	prev    []uint16 // Index to index of prev. occurance of same hash.
	hashVal []int    // Index to hash value at this index.
	val     int      // Current hash value.

	// Only used when HASH_SAME_HASH is true
	// Fields with similar purpose as the above hash, but for the second
	// hash with a value that is calculated differently.
	head2    []int    // Hash value to index of its most recent occurance.
	prev2    []uint16 // Index to index of prev. occurance of same hash.
	hashVal2 []int    // Index to hash value at this index.
	val2     int      // Current hash value.

	// Only used when HASH_SAME is true
	same []uint16 // Amount of repetitions of same byte after this.
}

// Allocates and initializes all fields of Hash.
func newHash(a, b byte) (h hash) {
	h.head = make([]int, 65536)
	h.prev = make([]uint16, WINDOW_SIZE)
	h.hashVal = make([]int, WINDOW_SIZE)
	for i := 0; i < 65536; i++ {
		h.head[i] = -1 // -1 indicates no head so far.
	}
	for i := uint16(0); i < WINDOW_SIZE; i++ {
		h.prev[i] = i // If prev[j] == j, then prev[j] is uninitialized.
		h.hashVal[i] = -1
	}

	if HASH_SAME {
		h.same = make([]uint16, WINDOW_SIZE)
	}

	if HASH_SAME_HASH {
		h.head2 = make([]int, 65536)
		h.prev2 = make([]uint16, WINDOW_SIZE)
		h.hashVal2 = make([]int, WINDOW_SIZE)
		for i := 0; i < 65536; i++ {
			h.head2[i] = -1
		}
		for i := uint16(0); i < WINDOW_SIZE; i++ {
			h.prev2[i] = i
			h.hashVal2[i] = -1
		}
	}

	h.warmup(a, b)
	return h
}

// Update the sliding hash value with the given byte. All calls to this function
// must be made on consecutive input characters. Since the hash value exists out
// of multiple input bytes, a few warmups with this function are needed initially.
func (h *hash) updateValue(c byte) {
	h.val = ((h.val << HASH_SHIFT) ^ int(c)) & HASH_MASK
}

// Updates the hash values based on the current position in the array. All calls
// to this must be made for consecutive bytes.
func (h *hash) update(slice []byte, pos, end int) {
	hPos := pos & WINDOW_MASK

	var hashValue byte
	if pos+MIN_MATCH <= end {
		hashValue = slice[pos+MIN_MATCH-1]
	}
	h.updateValue(hashValue)
	h.hashVal[hPos] = h.val
	if h.head[h.val] != -1 && h.hashVal[h.head[h.val]] == h.val {
		h.prev[hPos] = uint16(h.head[h.val])
	} else {
		h.prev[hPos] = uint16(hPos)
	}
	h.head[h.val] = hPos

	if HASH_SAME {
		// Update "same".
		var amount int
		if h.same[(pos-1)&WINDOW_MASK] > 1 {
			amount = int(h.same[(pos-1)&WINDOW_MASK]) - 1
		}
		for pos+amount+1 < end &&
			slice[pos] == slice[pos+amount+1] && amount < 0xFFFF {
			amount++
		}
		h.same[hPos] = uint16(amount)
	}

	if HASH_SAME_HASH {
		h.val2 = int((h.same[hPos]-MIN_MATCH)&255) ^ h.val
		h.hashVal2[hPos] = h.val2
		if h.head2[h.val2] != -1 && h.hashVal2[h.head2[h.val2]] == h.val2 {
			h.prev2[hPos] = uint16(h.head2[h.val2])
		} else {
			h.prev2[hPos] = uint16(hPos)
		}
		h.head2[h.val2] = hPos
	}
}

// Prepopulates hash:
// Fills in the initial values in the hash, before Update can be used
// correctly.
func (h *hash) warmup(a, b byte) {
	h.updateValue(a)
	h.updateValue(b)
}
