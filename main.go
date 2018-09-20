/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 12-11-2017
 * |
 * | File Name:     main.go
 * +===============================================
 */

package main

import (
	"fmt"

	"github.com/I1820/link/mqtt"
)

func main() {
	fmt.Println("18.20 at Sep 07 2016 7:20 IR721")

	mqtt.New().Run()
	for {
	}
}
