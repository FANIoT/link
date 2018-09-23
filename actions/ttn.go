/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 24-09-2018
 * |
 * | File Name:     ttn.go
 * +===============================================
 */

package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/I1820/types"
	"github.com/gobuffalo/buffalo"
	"github.com/polydawn/refmt/cbor"
	"github.com/sirupsen/logrus"
)

// TTNRequest is a data format that ttn http integration module sends
type TTNRequest struct {
	AppID          string `json:"app_id"`
	DevID          string `json:"dev_id"`
	HardwareSerial string `json:"hardware_serial"`
	Port           int    `json:"port"`
	Counter        int    `json:"counter"`
	PayloadRaw     []byte `json:"payload_raw"`

	Metadata struct {
		Time time.Time `json:"time"`
	} `json:"metadata"`
}

// TTNHandler provides an endpoint for TheThingsNetwork HTTP integration
// https://www.thethingsnetwork.org/docs/applications/http/
// This function is mapped to the path POST /ttn/{project_id}
// TODO it tries to decode given data with CBOR but is must improve to support
// more protocol and better configuration
func TTNHandler(c buffalo.Context) error {
	projectID := c.Param("project_id")

	var rq TTNRequest
	if err := c.Bind(&rq); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	coreApp.Logger.WithFields(logrus.Fields{
		"component": "ttn service",
	}).Infof("Incoming data from %s : %s with pid: %s", rq.AppID, rq.DevID, projectID)

	var states map[interface{}]interface{}
	if err := cbor.Unmarshal(rq.PayloadRaw, &states); err != nil {
		coreApp.Logger.WithFields(logrus.Fields{
			"component": "ttn service",
		}).Errorf("Incoming data from %s : %s with pid: %s is not a valid cbor (%q)", rq.AppID, rq.DevID, projectID, rq.PayloadRaw)
		return c.Render(http.StatusOK, r.JSON(true))
	}

	for name, value := range states {
		state := types.State{
			Raw: value,
			At:  rq.Metadata.Time,
			// TODO ThingID:
			Asset: fmt.Sprintf("%v", name), // convert anything to string (is there any better way?)
		}
		fmt.Println(state)
	}

	return c.Render(http.StatusOK, r.JSON(true))
}
