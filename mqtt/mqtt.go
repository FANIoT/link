/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 20-09-2018
 * |
 * | File Name:     mqtt.go
 * +===============================================
 */

package mqtt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/I1820/link/core"
	"github.com/I1820/types"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/gobuffalo/envy"
	"github.com/sirupsen/logrus"
)

func init() {
	rand.Seed(time.Now().UnixNano())

}

// Service of link component
// this service provide a way for users to send their data
// based on mqtt.
type Service struct {
	cli paho.Client
	app *core.Application
}

// New creates new mqtt service
func New() *Service {
	s := Service{}
	s.app = core.New()

	return &s
}

// handler handles incoming mqtt messages for following topic
// /things/{thing_id}/state
func (s *Service) handler(client paho.Client, message paho.Message) {
	thingID := strings.Split(message.Topic(), "/")[1]

	var states map[string]struct {
		At    time.Time
		Value interface{}
	}

	if err := json.Unmarshal(message.Payload(), &states); err != nil {
		s.app.Logger.WithFields(logrus.Fields{
			"component": "mqtt service",
			"topic":     message.Topic(),
		}).Errorf("Marshal error %s: %s", err, message.Payload())
		return
	}
	s.app.Logger.WithFields(logrus.Fields{
		"component": "mqtt service",
		"topic":     message.Topic(),
	}).Infof("Marshal on %v", states)

	for name, state := range states {
		s.app.Data(types.State{
			Raw:     state.Value,
			At:      state.At,
			ThingID: thingID,
			Asset:   name,
		})
	}
}

// Run runs mqtt service
func (s *Service) Run() error {
	// Create an MQTT client for mqtt service
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
	opts.AddBroker(envy.Get("USR_BROKER_URL", "tcp://127.0.0.1:1883"))
	opts.SetClientID(fmt.Sprintf("I1820-mqs-link-%d", rand.Intn(1024)))
	opts.SetOnConnectHandler(func(client paho.Client) {
		if t := s.cli.Subscribe("$share/i1820-link/things/+/state", 0, s.handler); t.Error() != nil {
			s.app.Logger.Fatalf("MQTT subscribe error: %s", t.Error())
		}
	})
	s.cli = paho.NewClient(opts)

	// Connect to the MQTT Server.
	if t := s.cli.Connect(); t.Wait() && t.Error() != nil {
		return t.Error()
	}
	s.app.Run()

	return nil
}
