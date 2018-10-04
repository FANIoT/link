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
	"sync"
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
// Application is used with link services in order to process
// Pipeline of application consists of following stages
// - Project Stage
// - Decode Stage
// - Insert Stage
type Application struct {
	cli paho.Client

	Logger *logrus.Logger

	session *mgo.Client
	db      *mgo.Database

	// pipeline channels
	projectStream chan *types.State
	decodeStream  chan *types.State
	insertStream  chan *types.State

	// in order to close the pipeline nicely
	projectCloseChan   chan struct{}  // project stage sends one value to this channel on its return
	decodeCloseChan    chan struct{}  // decode stage sends one value to this channel on its return
	insertCloseCounter sync.WaitGroup // count number of insert stages so `Exit` can wait for all of them

	IsRun bool
}

// New creates new application. this function does not create mqtt client.
// it creates mongodb session instance
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

// Run runs application. this function creates and connects mqtt client.
func (a *Application) Run() {
	// create close channels here so we can run and stop single
	// application many times
	a.projectCloseChan = make(chan struct{}, 1)
	a.decodeCloseChan = make(chan struct{}, 1)

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
	opts.AddBroker(envy.Get("SYS_BROKER_URL", "tcp://127.0.0.1:18083"))
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
		a.insertCloseCounter.Add(1)
	}

	a.IsRun = true
}

// Exit closes mqtt connection then closes all channels and return from all pipeline stages
func (a *Application) Exit() {
	a.IsRun = false

	// disconnect waiting time in milliseconds
	var quiesce uint = 10
	a.cli.Disconnect(quiesce)

	// close project stream
	close(a.projectStream)

	// all channels are going to close
	// so we are waiting for them
	a.insertCloseCounter.Wait()
}

// Data sends incoming data into application for futher processing
// incomming data must have raw, at, thingid and assets section of data
// please note that this function is a blocking function.
func (a *Application) Data(s types.State) error {
	if !a.IsRun {
		return fmt.Errorf("You cann't pass data into application when it is not running")
	}
	if s.Raw == nil || s.At.IsZero() {
		return fmt.Errorf("Raw and At must not be zero")
	}

	if s.ThingID == "" || s.Asset == "" {
		return fmt.Errorf("ThingID and Asset must not be empty")
	}

	a.projectStream <- &s
	return nil
}
