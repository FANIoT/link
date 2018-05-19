/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 17-11-2017
 * |
 * | File Name:     main.go
 * +===============================================
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aiotrc/downlink/encoder"
	"github.com/aiotrc/downlink/lora"
	pmclient "github.com/aiotrc/pm/client"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// Config represents main configuration
var Config = struct {
	Broker struct {
		URL string `default:"127.0.0.1:1883" env:"broker_url"`
	}
	Encoder struct {
		Host string `default:"127.0.0.1" env:"encoder_host"`
	}
	PM struct {
		URL string `default:"http://127.0.0.1:8080" env:"pm_url"`
	}
}{}

var pm pmclient.PM
var cli *client.Client

// handle registers apis and create http handler
func handle() http.Handler {
	r := gin.Default()

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "404 Not Found"})
	})

	r.Use(gin.ErrorLogger())

	api := r.Group("/api")
	{
		api.GET("/about", aboutHandler)

		api.POST("/send", sendHandler)
	}

	return r
}

func main() {
	// Load configuration
	if err := configor.Load(&Config, "config.yml"); err != nil {
		panic(err)
	}

	pm = pmclient.New(Config.PM.URL)

	// Create an MQTT client
	cli = client.New(&client.Options{
		ErrorHandler: func(err error) {
			log.WithFields(log.Fields{
				"component": "downlink",
			}).Errorf("MQTT Client %s", err)
		},
	})
	defer cli.Terminate()

	// Connect to the MQTT Server.
	if err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  Config.Broker.URL,
		ClientID: []byte(fmt.Sprintf("isrc-push-%d", rand.Int63())),
	}); err != nil {
		log.Fatalf("MQTT session %s: %s", Config.Broker.URL, err)
	}
	fmt.Printf("MQTT session %s has been created\n", Config.Broker.URL)

	fmt.Println("Downlink AIoTRC @ 2018")

	r := handle()

	srv := &http.Server{
		Addr:    ":1373",
		Handler: r,
	}

	go func() {
		fmt.Printf("Downlink Listen: %s\n", srv.Addr)
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Listen Error:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Downlink Shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown Error:", err)
	}
}

func aboutHandler(c *gin.Context) {
	c.String(http.StatusOK, "18.20 is leaving us")
}

func sendHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var r sendReq
	if err := c.BindJSON(&r); err != nil {
		return
	}

	p, err := pm.GetThingProject(r.ThingID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	encoder := encoder.New(fmt.Sprintf("http://%s:%s", Config.Encoder.Host, p.Runner.Port))

	raw, err := encoder.Encode(r.Data, r.ThingID)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	b, err := json.Marshal(lora.TxMessage{
		FPort:     r.FPort,
		Data:      raw,
		Confirmed: r.Confirmed,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err := cli.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte(fmt.Sprintf("application/%s/node/%s/tx", p.Name, r.ThingID)),
		Message:   b,
	}); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, raw)
}
