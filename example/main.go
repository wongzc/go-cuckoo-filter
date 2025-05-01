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
	c := cuckoofilter.New(10, 0.1)
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
				fmt.Printf("Deleted %s from filter.", str)
			}
		}
	}
}
