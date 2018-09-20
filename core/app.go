/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 02-08-2018
 * |
 * | File Name:     app.go
 * +===============================================
 */

package core

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/I1820/types"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/gobuffalo/envy"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Application is a main part of link component that consists of
// mqtt client and protocols that provide information for mqtt connectivity
type Application struct {
	cli paho.Client

	Logger *logrus.Logger

	session *mgo.Client
	db      *mgo.Database

	// pipeline channels
	projectStream chan *types.State
	decodeStream  chan *types.State
	insertStream  chan *types.State
}

// New creates new application. this function creates mqtt client
func New() *Application {
	a := Application{}

	a.Logger = logrus.New()

	// Create a mongodb connection
	url := envy.Get("DB_URL", "mongodb://127.0.0.1:27017")
	session, err := mgo.NewClient(url)
	if err != nil {
		a.Logger.Fatalf("DB new client error: %s", err)
	}
	a.session = session

	// pipeline channels
	a.projectStream = make(chan *types.State)
	a.decodeStream = make(chan *types.State)
	a.insertStream = make(chan *types.State)

	return &a
}

// Run runs application. this function connects mqtt client and then register its topic
func (a *Application) Run() {
	// Create an MQTT client
	/*
		Port: 1883
		CleanSession: True
		Order: True
		KeepAlive: 30 (seconds)
		ConnectTimeout: 30 (seconds)
		MaxReconnectInterval 10 (minutes)
		AutoReconnect: True
	*/
	opts := paho.NewClientOptions()
	opts.AddBroker(envy.Get("SYS_BROKER_URL", "tcp://127.0.0.1:1883"))
	opts.SetClientID(fmt.Sprintf("I1820-link-%d", rand.Intn(1024)))
	opts.SetOrderMatters(false)
	a.cli = paho.NewClient(opts)

	// Connect to the MQTT Server.
	if t := a.cli.Connect(); t.Wait() && t.Error() != nil {
		a.Logger.Fatalf("MQTT session error: %s", t.Error())
	}

	// Connect to the mongodb
	if err := a.session.Connect(context.Background()); err != nil {
		a.Logger.Fatalf("DB connection error: %s", err)
	}
	a.db = a.session.Database("i1820")

	// pipeline stages
	for i := 0; i < runtime.NumCPU(); i++ {
		go a.projectStage()
		go a.decodeStage()
		go a.insertStage()
	}
}

// Data sends incoming data into application for futher processing
// incomming data must have raw, at, thingid and assets section of data
// please note that this function is a blocking function.
func (a *Application) Data(s types.State) error {
	if s.Raw == nil || s.At.IsZero() {
		return fmt.Errorf("Raw and At must not be zero")
	}

	if s.ThingID == "" || s.Asset == "" {
		return fmt.Errorf("ThingID and Asset must not be empty")
	}

	a.projectStream <- &s
	return nil
}
