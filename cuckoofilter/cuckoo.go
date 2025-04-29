package cuckoofilter

import (
	"crypto/sha1"
)

type Cuckoo struct {
	buckets []bucket
	m       uint // buckets
	b       uint // entries per bucket
	f       uint // fingerprint length
	n       uint // filter capacity
}

type bucket []fingerprint
type fingerprint []byte

var hasher = sha1.New()

func New(n uint, fp float64) *Cuckoo {
	b := uint(4)
	f := fingerprintLength(b, fp)
	m := nextPower(n / f * 8)
	buckets := make([]bucket, m)
	for i := uint(0); i < m; i++ {
		buckets[i] = make(bucket, b)
	}
	return &Cuckoo{
		buckets: buckets,
		m:       m,
		b:       b,
		f:       f,
		n:       n,
	}
}
