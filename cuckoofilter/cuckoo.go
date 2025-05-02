package cuckoofilter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/cespare/xxhash/v2"
	"math"
	"math/rand"
)

type Cuckoo struct {
	buckets           []bucket
	BucketCount       uint64
	BucketSize        uint64
	FingerPrintLength uint64
	MaxRetries        int
}

type bucket []fingerprint
type fingerprint []byte

func (b bucket) nextIndex() (int, error) {
	for i, f := range b {
		if f == nil {
			return i, nil
		}
	}
	return -1, errors.New("bucket full")
}

func (b bucket) contains(f fingerprint) (int, bool) {
	for i, x := range b {
		if bytes.Equal(x, f) {
			return i, true
		}
	}
	return -1, false
}

func nextPower(i uint64) uint64 {
	// get min power of 2 that is larger, to use bitwise masking
	if i == 0 {
		return 1
	}
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

func New(itemCount uint64, accuracy float64, bucketSize uint64) *Cuckoo {
	maxLoadFactor := map[uint64]uint64{
		1: 50,
		2: 84,
		4: 95,
		8: 98,
	} // max load factor under each bucket size by Fan et al.
	lf, ok := maxLoadFactor[bucketSize]
	if !ok {
		panic("unsupported bucket size, use 1, 2, 4, or 8")
	}
	fingerPrintLength := uint64(math.Ceil(math.Log(2 * float64(bucketSize) / accuracy)))
	if fingerPrintLength < 1 {
		fingerPrintLength = 1
	}
	bucketCount := nextPower(itemCount * 100 / (bucketSize * lf))
	buckets := make([]bucket, bucketCount)
	for i := uint64(0); i < bucketCount; i++ {
		buckets[i] = make(bucket, bucketSize)
	}

	maxRetries := int(10*math.Log2(float64(itemCount))) + 1
	return &Cuckoo{
		buckets:           buckets,
		BucketCount:       bucketCount,
		BucketSize:        bucketSize,
		FingerPrintLength: fingerPrintLength, // in bits
		MaxRetries:        maxRetries,
	}
}

func (c *Cuckoo) hashes(data string) (uint64, uint64, fingerprint) {
	h := hash([]byte(data))
	f := fingerprintBits(h, c.FingerPrintLength)
	i1 := h
	i2 := i1 ^ hash(f)
	return i1, i2, fingerprint(f)
}

func hash(data []byte) uint64 {
	return xxhash.Sum64(data)
}

func (c *Cuckoo) Set(data string) error {
	i1, i2, f := c.hashes(data)
	mask := c.BucketCount - 1
	b1 := c.buckets[i1&mask]
	if i, err := b1.nextIndex(); err == nil {
		b1[i] = f
		return nil
	}

	b2 := c.buckets[i2&mask]
	if i, err := b2.nextIndex(); err == nil {
		b2[i] = f
		return nil
	}

	i := i1
	for r := 0; r < c.MaxRetries; r++ {
		index := i % c.BucketCount
		entryIndex := rand.Intn(int(c.BucketSize))
		f, c.buckets[index][entryIndex] = c.buckets[index][entryIndex], f
		i = i ^ hash(f)
		b := c.buckets[i&mask]
		if idx, err := b.nextIndex(); err == nil {
			b[idx] = f
			return nil
		}
	}
	return errors.New("Cuckoo filter full")
}

func (c *Cuckoo) Del(needle string) {
	i1, i2, f := c.hashes(needle)
	mask := c.BucketCount - 1
	b1 := c.buckets[i1&mask]
	if ind, ok := b1.contains(f); ok {
		b1[ind] = nil
		return
	}

	b2 := c.buckets[i2&mask]
	if ind, ok := b2.contains(f); ok {
		b2[ind] = nil
		return
	}
}

func (c *Cuckoo) Get(needle string) bool {
	i1, i2, f := c.hashes(needle)
	mask := c.BucketCount - 1
	_, b1 := c.buckets[i1&mask].contains(f)
	_, b2 := c.buckets[i2&mask].contains(f)
	return b1 || b2
}

func fingerprintBits(h uint64, bitLen uint64) []byte {
	byteLen := (bitLen + 7) / 8
	full := make([]byte, 8)
	binary.BigEndian.PutUint64(full, h)

	fp := make([]byte, byteLen)
	copy(fp, full[:byteLen])

	// Zero out excess bits in the last byte
	excessBits := (8 * byteLen) - bitLen
	if excessBits > 0 {
		fp[byteLen-1] &= (1 << (8 - excessBits)) - 1
	}

	allZero := true
	for _, b := range fp {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero { // ensure it is not 0, else maybe deemed as empty
		fp[0] = 1
	}

	return fp
}
