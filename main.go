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
	"log"
	"os"
	"os/signal"

	"github.com/I1820/link/actions"
	"github.com/I1820/link/mqtt"
)

func main() {
	fmt.Println("18.20 at Sep 07 2016 7:20 IR721")

	// buffalo http service
	go func() {
		app := actions.App()
		if err := app.Serve(); err != nil {
			log.Fatalf("Buffalo Service failed with %s", err)
		}
	}()
	// non-http services
	if err := mqtt.New().Run(); err != nil {
		log.Fatalf("MQTT Service failed with %s", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	<-sigc
}
