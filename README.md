# Cuckoo Filter in Go
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)  
This project implements a [Cuckoo Filter](https://www.pdl.cmu.edu/PDL-FTP/FS/cuckoo-conext2014.pdf), inspired by the paper *"Cuckoo Filter: Practically Better Than Bloom"* by Bin Fan, David G. Andersen, and Michael Kaminsky (CoNEXT 2014).

## Installation

```bash
go get github.com/wongzc/go-cuckoo-filter/cuckoofilter
```

## What is a Cuckoo Filter?

A Cuckoo Filter is a probabilistic data structure used for set membership tests, similar to a Bloom Filter—but with added support for deletions, better lookup performance, and improved space efficiency when the desired false positive rate (FPR) is under 3%.

Instead of storing the full item, it stores a short fingerprint of each item in one of two possible buckets. Like Cuckoo Hashing, it may relocate existing entries to make space for new ones.


## Advantages

- Supports **dynamic add and delete** operations.
- **Faster lookups** than Bloom Filters, even at 95% capacity.
- **Simpler to implement** than Bloom Filter variants that support deletion.
- **More space-efficient** than Bloom Filters when FPR < 3%.


## Theory

### Cuckoo Hash Table Basics

- Uses 1 hash functions.
- Each item has **two candidate buckets**: `h1(x)` and `h2(x) = h1(x) ⊕ hash(fingerprint(x))`.
- If both buckets are full, evict an existing item (Cuckooing), and recursively reinsert the evicted one.
- Each bucket may store multiple entries (typically 2–8).
- Average insertion time is **O(1)**.

### Fingerprint Storage

- Only stores a short **fingerprint** of each item.
- Fingerprint size is based on the **desired false positive rate (ε)**.
- Insertion uses **partial-key Cuckoo hashing**, relying only on the fingerprint.

### Lookup

- Compute fingerprint `f = fingerprint(x)`.
- Check if `f` exists in either `h1(x)` or `h2(x)`.
- Guarantees **no false negatives** unless an insertion failed due to full capacity.

### Deletion

- Check both buckets for the fingerprint.
- Remove it if found.
- No risk of **false deletion**:
  - If multiple items share a fingerprint, their alternate bucket will still retain one valid copy.
  - **FPR remains unchanged** after deletion.


## Formulas Used

### Bucket Size

$$
b = 2, 4, 8
$$
- Smaller `b` → faster lookup.
- Larger `b` → higher load factor (more space efficiency).

### Minimum Fingerprint Length
$$
f = \left\lceil \log_2 (\frac{2b}{\varepsilon})\right\rceil
$$

- `ε`: false positive rate
- `b`: bucket size
- `f`: bits per fingerprint

### Number of Bucket
$$
m= \left\lceil \frac{n}{\alpha \cdot b} \right\rceil
$$
- `n`: number of items
- `α`: expected load factor (e.g. 0.84 for b=4)
- `b`: bucket size
- Round `m` up to the next power of 2 for fast modulo operations using bit masking.

### Maximum Retry Limit
```math
\text{max\_retries}=10\cdot\log_2(n)
```
- `n`: number of items
- Used to cap the number of relocation attempts during insertion.

## Example Usage

```go
package main

import (
    "fmt"
    "github.com/wongzc/go-cuckoo-filter/cuckoofilter"
)

func main() {
    cf := cuckoofilter.New(1000, 0.01, 4)

    cf.Set("hello")

    fmt.Println(cf.Get("hello")) // true
    fmt.Println(cf.Get("world")) // probably false

    cf.Del("hello")
    fmt.Println(cf.Get("hello")) // false
}
```