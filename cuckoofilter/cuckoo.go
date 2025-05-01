package cuckoofilter

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
)

type Cuckoo struct {
	buckets           []bucket
	bucketCount       uint
	bucketSize        uint
	fingerPrintLength uint
	maxRetries        int
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

func fingerprintLength(bucketSize uint, accuracy float64) uint {
	f := uint(math.Ceil(math.Log(2 * float64(bucketSize) / accuracy)))
	if f < 1 {
		return 1
	}
	return f
}

func nextPower(i uint) uint {
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

var hasher = sha1.New()

func New(itemCount uint, accuracy float64) *Cuckoo {
	bucketSize := uint(4)
	f := fingerprintLength(bucketSize, accuracy)
	bucketCount := nextPower(itemCount / f * 8)
	buckets := make([]bucket, bucketCount)
	for i := uint(0); i < bucketCount; i++ {
		buckets[i] = make(bucket, bucketSize)
	}
	return &Cuckoo{
		buckets:           buckets,
		bucketCount:       bucketCount,
		bucketSize:        bucketSize,
		fingerPrintLength: f,
		maxRetries:        int(10*math.Log2(float64(itemCount))) + 1,
	}
}

func (c *Cuckoo) hashes(data string) (uint, uint, fingerprint) {
	h := hash([]byte(data))
	f := h[0:c.fingerPrintLength]
	i1 := uint(binary.BigEndian.Uint32(h))
	i2 := i1 ^ uint(binary.BigEndian.Uint32(hash(f)))
	return i1, i2, fingerprint(f)
}

func hash(data []byte) []byte {
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)
	hasher.Reset()
	return hash
}

func (c *Cuckoo) Set(data string) {
	i1, i2, f := c.hashes(data)
	b1 := c.buckets[i1%c.bucketCount]
	if i, err := b1.nextIndex(); err == nil {
		b1[i] = f
		return
	}

	b2 := c.buckets[i2%c.bucketCount]
	if i, err := b2.nextIndex(); err == nil {
		b2[i] = f
		return
	}

	i := i1
	for r := 0; r < c.maxRetries; r++ {
		index := i % c.bucketCount
		entryIndex := rand.Intn(int(c.bucketSize))
		f, c.buckets[index][entryIndex] = c.buckets[index][entryIndex], f
		i = i ^ uint(binary.BigEndian.Uint32(hash(f)))
		b := c.buckets[i%c.bucketCount]
		if idx, err := b.nextIndex(); err == nil {
			b[idx] = f
			return
		}
	}
	panic("cuckoo filter full")
}

func (c *Cuckoo) Del(needle string) {
	i1, i2, f := c.hashes(needle)
	b1 := c.buckets[i1%c.bucketCount]
	if ind, ok := b1.contains(f); ok {
		b1[ind] = nil
		return
	}

	b2 := c.buckets[i2%c.bucketCount]
	if ind, ok := b2.contains(f); ok {
		b2[ind] = nil
		return
	}
}

func (c *Cuckoo) Get(needle string) bool {
	i1, i2, f := c.hashes(needle)
	_, b1 := c.buckets[i1%c.bucketCount].contains(f)
	_, b2 := c.buckets[i2%c.bucketCount].contains(f)
	return b1 || b2
}
