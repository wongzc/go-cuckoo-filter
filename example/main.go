package main

import (
	"bufio"
	"fmt"
	"github.com/wongzc/go-cuckoo-filter/cuckoofilter"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var item_count uint64
	var accuracy float64
	var bucketSize uint64
	fmt.Println("Enter estimated item count (integer), acceptable false positive rate (0<x<1), bucket size (1,2,4,8)")
	_, err := fmt.Scan(&item_count, &accuracy, &bucketSize)
	if err != nil || accuracy <= 0 || accuracy >= 1 || bucketSize !=2 && bucketSize!=4 && bucketSize!=8 && bucketSize!=1{
		fmt.Println("Invalid input")
		return
	}
	c := cuckoofilter.New(item_count, accuracy, bucketSize)

	fmt.Printf("Using %d buckets (size: %d). Fingerprint length: %d. Max retries: %d.\n", c.BucketCount, c.BucketSize, c.FingerPrintLength, c.MaxRetries)
	for {
		fmt.Print("[s=set, g=get, d=delete, x=exit]: ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if cmd == "x" {
			break
		} else if cmd == "s" || cmd == "g" || cmd == "d" {
			fmt.Print("Enter string: ")
			str, _ := reader.ReadString('\n')
			str = strings.TrimSpace(str)
			if cmd == "s" {
				c.Set(str)
			} else if cmd == "g" {
				if c.Get(str) {
					fmt.Println("Probably in set.")
				} else {
					fmt.Println("Definitely not in set")
				}
			} else {
				c.Del(str)
				fmt.Printf("Deleted %s from filter.\n", str)
			}
		}
	}
}
