package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"

	pmclient "github.com/aiotrc/pm/client"
	"github.com/aiotrc/uplink/decoder"
	"github.com/aiotrc/uplink/lora"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

// Config represents main configuration
var Config = struct {
	DB struct {
		URL string `default:"mongodb://127.0.0.1" env:"db_url"`
	}
	Broker struct {
		URL string `default:"127.0.0.1:1883" env:"broker_url"`
	}
	Decoder struct {
		Host string `default:"127.0.0.1" env:"decoder_host"`
	}
	PM struct {
		URL string `default:"http://127.0.0.1:8080" env:"pm_url"`
	}
}{}

// App creates configured mqtt application
func App() {
	// Load configuration
	if err := configor.Load(&Config, "config.yml"); err != nil {
		panic(err)
	}

	// Create a Mongo Session
	session, err := mgo.Connect(context.Background(), Config.DB.URL, nil)
	if err != nil {
		log.Fatalf("Mongo session %s: %v", Config.DB.URL, err)
	}

	// Create an MQTT client
	cli := client.New(&client.Options{
		ErrorHandler: func(err error) {
			log.WithFields(log.Fields{
				"component": "uplink",
			}).Errorf("MQTT Client %s", err)
		},
	})
	defer cli.Terminate()

	// Connect to the MQTT Server.
	if err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  Config.Broker.URL,
		ClientID: []byte(fmt.Sprintf("isrc-uplink-%d", rand.Int63())),
	}); err != nil {
		log.Fatalf("MQTT session %s: %s", Config.Broker.URL, err)
	}
	fmt.Printf("MQTT session %s has been created\n", Config.Broker.URL)

	// PM
	pm := pmclient.New(Config.PM.URL)

	// LoRa error collection
	ce := session.Database("isrc").Collection("lora")

	// Data collection
	cd := session.Database("isrc").Collection("data")
	indx, err := cd.Indexes().CreateMany(
		context.Background(),
		mgo.IndexModel{
			Keys: bson.NewDocument(
				bson.EC.Int32("timestamp", 1),
			),
		},
		mgo.IndexModel{
			Keys: bson.NewDocument(
				bson.EC.Int32("thingid", 1),
				bson.EC.Int32("timestamp", 1),
			),
		},
		mgo.IndexModel{
			Keys: bson.NewDocument(
				bson.EC.String("data._location", "2dsphere"),
			),
		},
	)
	if err != nil {
		log.Fatalf("Create index %v", err)
	}
	fmt.Printf("MongoDB \"data\" collection indexes: %v\n", indx)

	// Subscribe to topics
	err = cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			// https://docs.loraserver.io/use/getting-started/
			&client.SubReq{
				TopicFilter: []byte("application/+/node/+/error"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					var m lora.ErrorMessage
					if err := json.Unmarshal(message, &m); err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("JSON Unmarshal: %s", err)
						return
					}
					log.WithFields(log.Fields{
						"component": "uplink",
					}).Info(m)
					if _, err := ce.InsertOne(context.Background(), &struct {
						Error     string
						Timestamp time.Time
						Type      string
						Project   string
						FCnt      int
					}{
						Error:     m.Error,
						Timestamp: time.Now(),
						Project:   m.ApplicationName,
						Type:      m.Type,
						FCnt:      m.FCnt,
					}); err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("Mongo insert: %s\n", err)
						return
					}

				},
			},
			&client.SubReq{
				TopicFilter: []byte("application/+/node/+/rx"),
				QoS:         mqtt.QoS0,
				Handler: func(topicName, message []byte) {
					var m lora.RxMessage
					if err := json.Unmarshal(message, &m); err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("JSON Unmarshal: %s", err)
						return
					}
					log.WithFields(log.Fields{
						"component": "uplink",
					}).Info(m)

					var bdoc interface{}

					// Find thing
					p, err := pm.GetThingProject(m.DevEUI)
					if err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("PM GetThingProject: %s", err)
						return
					}
					// TODO: thing activation
					/*
						if !t.Status {
							return
						}
					*/

					defer func() {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Info("Insert into databse")

						if _, err := cd.InsertOne(context.Background(), &struct {
							Raw       []byte
							Data      interface{}
							Timestamp time.Time
							ThingID   string
							RxInfo    []lora.RxInfo
							TxInfo    lora.TxInfo
							Project   string
						}{
							Raw:       m.Data,
							Data:      bdoc,
							Timestamp: time.Now(),
							ThingID:   m.DevEUI,
							RxInfo:    m.RxInfo,
							TxInfo:    m.TxInfo,
							Project:   p.Name,
						}); err != nil {
							log.WithFields(log.Fields{
								"component": "uplink",
							}).Errorf("Mongo insert: %s\n", err)
							return
						}
					}()

					// Create decoder
					decoder := decoder.New(fmt.Sprintf("http://%s:%s", Config.Decoder.Host, p.Runner.Port))

					// Decode
					parsed, err := decoder.Decode(m.Data, m.DevEUI)
					if err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("Decode: %s", err)
						return
					}

					if err := json.Unmarshal([]byte(parsed), &bdoc); err != nil {
						log.WithFields(log.Fields{
							"component": "uplink",
						}).Errorf("Unmarshal JSON: %s\n %q", err, parsed)
						return
					}
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("MQTT subscription: %s", err)
	}
}
