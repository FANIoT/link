/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 01-10-2018
 * |
 * | File Name:     mqtt.go
 * +===============================================
 */

package actions

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FANIoT/link/pm"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
)

// VernemqAuthPlugin is an authentication plugin based vernemq webhooks
// see https://vernemq.com/docs/plugindevelopment/webhookplugins.html for more details
// This plugin validate things and their tokens.
type VernemqAuthPlugin struct{}

// VernemqRequest is a minimal request structure for its webhook request data
type VernemqRequest struct {
	PeerAddr     string `json:"peer_addr"`
	PeerPort     int    `json:"peer_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Mountpoint   string `json:"mountpoint"`
	ClientID     string `json:"client_id"`
	CleanSession bool   `json:"clean_session"`
	Topic        string `json:"topic"`
	Topics       []struct {
		Topic string `json:"topic"`
		QoS   int    `json:"qos"`
	} `json:"topics"`
}

// VernemqResponse is a minimal responses structure for its webhook response data
var (
	VernemqOKResponse = struct {
		Result string `json:"result"`
	}{
		Result: "ok",
	}

	VernemqErrorResponse = struct {
		Result map[string]string `json:"result"`
	}{
		Result: map[string]string{
			"error": "not ok",
		},
	}
)

// OnRegister is called when a new client connects to vernemq.
// It is better to authorize clients when they try to subscribe and publish
// data so this function always returns ok
func (VernemqAuthPlugin) OnRegister(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.JSON(VernemqOKResponse))
}

// OnSubscribe is called when a client tries to subscribe on a topic
func (VernemqAuthPlugin) OnSubscribe(c buffalo.Context) error {
	var req VernemqRequest
	if err := c.Bind(&req); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	if req.Mountpoint == "i1820" || req.Username == envy.Get("USR_BROKER_USER", "ella") {
		// let them pass, they have suffered enough
		c.Response().Header().Add("cache-control", fmt.Sprintf("max-age=%d", 3600*24)) // valid for one day
		return c.Render(http.StatusOK, r.JSON(VernemqOKResponse))
	}

	// everybody only can subscribe on one topic
	if len(req.Topics) != 1 {
		return c.Render(http.StatusOK, r.JSON(VernemqErrorResponse))
	}

	thingID := strings.Split(req.Topics[0].Topic, "/")[1]

	t, err := pm.ThingByID(c, thingID)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	for _, token := range t.Tokens {
		if token == req.Username {
			c.Response().Header().Add("cache-control", fmt.Sprintf("max-age=%d", 3600))
			return c.Render(http.StatusOK, r.JSON(VernemqOKResponse))
		}
	}

	return c.Render(http.StatusOK, r.JSON(VernemqErrorResponse))
}

// OnPublish is called when a client tries to publish data on a topic
func (VernemqAuthPlugin) OnPublish(c buffalo.Context) error {
	var req VernemqRequest
	if err := c.Bind(&req); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}

	if req.Mountpoint == "i1820" || req.Username == envy.Get("USR_BROKER_USER", "ella") {
		// let them pass, they have suffered enough
		c.Response().Header().Add("cache-control", fmt.Sprintf("max-age=%d", 3600*24)) // valid for one day
		return c.Render(http.StatusOK, r.JSON(VernemqOKResponse))
	}

	thingID := strings.Split(req.Topic, "/")[1]

	t, err := pm.ThingByID(c, thingID)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	for _, token := range t.Tokens {
		if token == req.Username {
			c.Response().Header().Add("cache-control", fmt.Sprintf("max-age=%d", 3600))
			return c.Render(http.StatusOK, r.JSON(VernemqOKResponse))
		}
	}

	return c.Render(http.StatusOK, r.JSON(VernemqErrorResponse))
}
