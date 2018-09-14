/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 02-08-2018
 * |
 * | File Name:     handler.go
 * +===============================================
 */

package app

import (
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
)

// mqttHandler creates mqtt handler for given protocol.
func (a *Application) mqttHandler(p Protocol) paho.MessageHandler {
	marshaler := p.Marshal

	return func(client paho.Client, message paho.Message) {
		// Data is created here for the first time.
		// It will be passed by reference from now on.
		d, err := marshaler(message.Payload())
		if err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
				"topic":     message.Topic(),
			}).Errorf("Marshal error %s", err)
			return
		}

		d.Protocol = p.Name()
		a.Logger.WithFields(logrus.Fields{
			"component": "link",
			"topic":     message.Topic(),
		}).Infof("Marshal on %v", d)
		a.projectStream <- &d
	}
}
