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

// Cache used by FindLongestMatch to remember previously found length/dist
// values.
// This is needed because the squeeze runs will ask these values multiple times for
// the same position.
// Uses large amounts of memory, since it has to remember the distance belonging
// to every possible shorter-than-the-best length (the so called "sublen" array).
type longestMatchCache struct {
	store  LZ77Store
	sublen []uint8
	active bool
}

// Initialize a LongestMatchCache.
func newCache(blockSize int) (lmc longestMatchCache) {
	lmc.store = make(LZ77Store, blockSize)
	// Rather large amount of memory.
	lmc.sublen = make([]uint8, CACHE_LENGTH*3*blockSize)
	lmc.active = true

	// length > 0 and dist 0 is invalid combination, which indicates on
	// purpose that this cache value is not filled in yet.
	for i := 0; i < blockSize; i++ {
		lmc.store[i].litLen = 1
	}

	return lmc
}

// Stores sublen array in the cache
func (lmc longestMatchCache) sublenToCache(sublen []uint16,
	pos int, length uint16) {
	var j, bestLength uint16

	if CACHE_LENGTH == 0 {
		return
	}

	cache := lmc.sublen[CACHE_LENGTH*pos*3:]
	if length < 3 {
		return
	}
	for i := uint16(3); i <= length; i++ {
		if i == length || sublen[i] != sublen[i+1] {
			cache[j*3] = uint8(i - 3)
			cache[j*3+1] = uint8(sublen[i])
			cache[j*3+2] = uint8(sublen[i] >> 8)
			bestLength = i
			j++
			if j >= CACHE_LENGTH {
				break
			}
		}
	}
	if j < CACHE_LENGTH {
		if bestLength != length {
			panic("couldn't find best length")
		}
		cache[(CACHE_LENGTH-1)*3] = uint8(bestLength - 3)
	} else {
		if bestLength > length {
			panic("impossible length")
		}
	}
	if bestLength != lmc.maxCachedSublen(pos) {
		panic("didn't cache sublen")
	}
}

// Extracts sublen array from the cache.
func (lmc longestMatchCache) cacheToSublen(pos int, length uint16, sublen []uint16) {
	if CACHE_LENGTH == 0 {
		return
	}

	if length < 3 {
		return
	}

	var prevLength uint16
	maxLength := lmc.maxCachedSublen(pos)
	cache := CACHE_LENGTH * pos * 3
	for j := 0; j < CACHE_LENGTH; j++ {
		length = uint16(lmc.sublen[cache+j*3]) + 3
		dist := uint16(lmc.sublen[cache+j*3+1]) + 256*uint16(lmc.sublen[cache+j*3+2])
		for i := prevLength; i <= length; i++ {
			sublen[i] = dist
		}
		if length == maxLength {
			break
		}
		prevLength = length + 1
	}
}

// Returns the length up to which could be stored in the cache.
func (lmc longestMatchCache) maxCachedSublen(pos int) uint16 {
	if CACHE_LENGTH == 0 {
		return 0
	}
	//cache := lmc.sublen[CACHE_LENGTH*pos*3:]
	cache := CACHE_LENGTH * pos * 3
	if lmc.sublen[cache+1] == 0 && lmc.sublen[cache+2] == 0 {
		//cache := lmc.sublen[CACHE_LENGTH*pos*3:]
		//if cache[1] == 0 && cache[2] == 0 {
		// No sublen cached.
		return 0
	}
	//return uint16(cache[(CACHE_LENGTH-1)*3]) + 3
	return uint16(lmc.sublen[cache+(CACHE_LENGTH-1)*3]) + 3
	//return uint16(lmc.sublen[(CACHE_LENGTH * (pos + 1) - 1) * 3]) + 3
}
