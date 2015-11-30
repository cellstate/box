/*
Copyright 2011 Google Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fsreverse

const windowSize = 64

// According to librsync/rollsum.h:
// "We should make this something other than zero to improve the
// checksum algorithm: tridge suggests a prime number."
// apenwarr: I unscientifically tried 0 and 7919, and they both ended up
// slightly worse than the librsync value of 31 for my arbitrary test data.
const charOffset = 31

const splitOnes = 13             // how many 1's as lower bits of the sum we consider a split
const splitSize = 1 << splitOnes // 2^13 = 8192 bytes on average

type RollSum struct {
	s1, s2 uint32
	window [windowSize]uint8
	wofs   int
}

func NewRollsum() *RollSum {
	return &RollSum{
		s1: windowSize * charOffset,
		s2: windowSize * (windowSize - 1) * charOffset,
	}
}

func (rs *RollSum) add(drop, add uint8) {
	rs.s1 += uint32(add) - uint32(drop)
	rs.s2 += rs.s1 - uint32(windowSize)*uint32(drop+charOffset)
}

func (rs *RollSum) Roll(ch byte) {
	rs.add(rs.window[rs.wofs], ch)
	rs.window[rs.wofs] = ch
	rs.wofs = (rs.wofs + 1) % windowSize
}

func (rs *RollSum) OnSplit() bool {
	return (rs.s2 & (splitSize - 1)) == ((^0) & (splitSize - 1))
}

//How many ones in the lower bits
func (rs *RollSum) Bits() int {
	bits := splitOnes
	rsum := rs.Digest()
	rsum >>= splitOnes
	for ; (rsum>>1)&1 != 0; bits++ {
		rsum >>= 1
	}
	return bits
}

func (rs *RollSum) Digest() uint32 {
	return (rs.s1 << 16) | (rs.s2 & 0xffff)
}
