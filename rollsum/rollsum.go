package rollsum

const windowSize = 64
const charOffset = 31

const blobBits = 13
const blobSize = 1 << blobBits // 8k

type RollSum struct {
	s1, s2 uint32
	window [windowSize]uint8
	wofs   int
}

func New() *RollSum {
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
	return (rs.s2 & (blobSize - 1)) == ((^0) & (blobSize - 1))
}

func (rs *RollSum) OnSplitWithBits(n uint32) bool {
	mask := (uint32(1) << n) - 1
	return rs.s2&mask == (^uint32(0))&mask
}

func (rs *RollSum) Bits() int {
	bits := blobBits
	rsum := rs.Digest()
	rsum >>= blobBits
	for ; (rsum>>1)&1 != 0; bits++ {
		rsum >>= 1
	}
	return bits
}

func (rs *RollSum) Digest() uint32 {
	return (rs.s1 << 16) | (rs.s2 & 0xffff)
}
