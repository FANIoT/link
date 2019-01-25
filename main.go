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
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/FANIoT/link/actions"
	"github.com/FANIoT/link/mqtt"
)

func main() {
	fmt.Println("18.20 at Sep 07 2016 7:20 IR721")

	var isHeadless = flag.Bool("headless", false, "Runs link in headless mode. In headless mode link just has its mqtt service")
	flag.Parse()

	if !*isHeadless {
		// buffalo http service
		go func() {
			app := actions.App()
			if err := app.Serve(); err != nil {
				log.Fatalf("Buffalo Service failed with %s", err)
			}
		}()
		runtime.Gosched()
		// we must wait for http service to go online before starting the
		// services. maybe in the future we change this behaviour.
		fmt.Println("Wait for http service to really start")
		time.Sleep(10 * time.Second)
	}
	// non-http services
	if err := mqtt.New().Run(); err != nil {
		log.Fatalf("MQTT Service failed with %s", err)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	<-sigc
}
