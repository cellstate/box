package rollsum

import (
	"bufio"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type span struct {
	from, to int64
	bits     int
	children []span
}

func UtilUpsertTestFile(t *testing.T, size int) *os.File {
	rnd := rand.New(rand.NewSource(5))
	fpath := filepath.Join(os.TempDir(), "test_cellstate_largefile")
	f, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(fpath)
			if err != nil {
				t.Fatal(err)
			}

			log.Printf("Writing %d random bytes to  %s...", size, fpath)
			total := 0
			for {
				n, err := f.Write([]byte{byte(rnd.Intn(256))})
				if err != nil {
					t.Fatal(err)
				}

				total += n
				if total >= size {
					f.Seek(0, 0)
					break
				}
			}

		} else {
			t.Fatal(err)
		}
	}

	return f
}

func TestSplitLargeFile(t *testing.T) {
	f := UtilUpsertTestFile(t, 100000)
	defer f.Close()
	log.Printf("%s", f.Name())

	buff := bufio.NewReader(f)
	spans := []span{}
	rs := New()
	n := int64(0)
	last := n
	for {
		b, err := buff.ReadByte()
		if err != nil {
			if err == io.EOF {
				if n != last {
					spans = append(spans, span{from: last, to: n})
				}
				break
			} else {
				log.Fatal(err)
			}
		}

		n++
		rs.Roll(b)
		if rs.OnSplit() {

			bits := rs.Bits()
			sliceFrom := len(spans)
			for sliceFrom > 0 && spans[sliceFrom-1].bits < bits {
				sliceFrom--
			}

			// nCopy := len(spans) - sliceFrom
			var children []span
			// if nCopy > 0 {
			// 	children = make([]span, nCopy)
			// 	nCopied := copy(children, spans[sliceFrom:])
			// 	if nCopied != nCopy {
			// 		panic("n wrong")
			// 	}
			// 	spans = spans[:sliceFrom]
			// }
			spans = append(spans, span{from: last, to: n, bits: bits, children: children})

			log.Printf("split at %d (after %d), bits=%d", n, n-last, bits)
			last = n
		}
	}

	log.Printf("\n")
	var dumpSpans func(s []span, indent int)
	dumpSpans = func(s []span, indent int) {
		in := strings.Repeat(" ", indent)
		for _, sp := range s {
			log.Printf("%sfrom=%d, to=%d (len %d) bits=%d\n", in, sp.from, sp.to, sp.to-sp.from, sp.bits)
			if len(sp.children) > 0 {
				dumpSpans(sp.children, indent+4)
			}
		}
	}

	dumpSpans(spans, 0)
}

func TestSum(t *testing.T) {
	var buf [100000]byte
	rnd := rand.New(rand.NewSource(4))
	for i := range buf {
		buf[i] = byte(rnd.Intn(256))
	}

	sum := func(offset, l int) uint32 {
		rs := New()
		for count := offset; count < l; count++ {
			rs.Roll(buf[count])
		}
		return rs.Digest()
	}

	sum1a := sum(0, len(buf))
	sum1b := sum(1, len(buf))
	sum2a := sum(len(buf)-windowSize*5/2, len(buf)-windowSize)
	sum2b := sum(0, len(buf)-windowSize)
	sum3a := sum(0, windowSize+3)
	sum3b := sum(3, windowSize+3)

	if sum1a != sum1b {
		t.Errorf("sum1a=%d sum1b=%d", sum1a, sum1b)
	}
	if sum2a != sum2b {
		t.Errorf("sum2a=%d sum2b=%d", sum2a, sum2b)
	}
	if sum3a != sum3b {
		t.Errorf("sum3a=%d sum3b=%d", sum3a, sum3b)
	}
}

func BenchmarkRollsum(b *testing.B) {
	const bufSize = 5 << 20
	buf := make([]byte, bufSize)
	for i := range buf {
		buf[i] = byte(rand.Int63())
	}

	b.ResetTimer()
	rs := New()
	splits := 0
	for i := 0; i < b.N; i++ {
		splits = 0
		for _, b := range buf {
			rs.Roll(b)
			if rs.OnSplit() {
				_ = rs.Bits()
				splits++
			}
		}
	}
	b.SetBytes(bufSize)
	b.Logf("num splits = %d; every %d bytes", splits, int(float64(bufSize)/float64(splits)))
}
