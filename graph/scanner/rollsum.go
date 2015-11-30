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

package scanner

import (
	"log"
)

const charOffset = 31            // something other then 0 (apparently a prime), improves checksum algo
const splitOnes = 13             // how many 1's as lower bits of the sum we consider a split
const splitSize = 1 << splitOnes // 2^13 = 8192 bytes on average
const windowSize = 64            // according to bups design document; chosen arbitrarly

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

//roll the window forward with one byte
func (rs *RollSum) Roll(ch byte) {
	rs.add(rs.window[rs.wofs], ch)
	rs.window[rs.wofs] = ch
	rs.wofs = (rs.wofs + 1) % windowSize
}

//returns true when the lowest bits of the
func (rs *RollSum) OnSplit() bool {

	res := (rs.s2 & (splitSize - 1)) == ((^0) & (splitSize - 1))
	if res == true {
		log.Printf("%b (%d): %t", rs.s2, splitSize, res)
	}

	return res
}

//How many ones in the lower bits of the sum
func (rs *RollSum) Bits() int {
	bits := splitOnes
	rsum := rs.Sum()
	rsum >>= splitOnes
	for ; (rsum>>1)&1 != 0; bits++ {
		rsum >>= 1
	}
	return bits
}

//get the checksum for the current window
func (rs *RollSum) Sum() uint32 {
	return (rs.s1 << 16) | (rs.s2 & 0xffff)
}
