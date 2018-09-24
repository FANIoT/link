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
	"strings"
	"time"

	"github.com/I1820/link/pm"
	"github.com/I1820/types"
	"github.com/I1820/types/connectivity"
	"github.com/gobuffalo/buffalo"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
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
	var thingID string

	var rq TTNRequest
	if err := c.Bind(&rq); err != nil {
		return c.Error(http.StatusBadRequest, err)
	}
	coreApp.Logger.WithFields(logrus.Fields{
		"component": "ttn service",
	}).Infof("Incoming data from %s @ %s with pid: %s", rq.DevID, rq.AppID, projectID)

	// return if there is no payload
	if rq.PayloadRaw == nil {
		return c.Render(http.StatusOK, r.JSON(true))
	}

	ts, err := pm.ThingsByProject(c, projectID)
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}
	for _, t := range ts {
		if c, ok := t.Connectivities["ttn"]; ok {
			var ttnC connectivity.TTN
			if err := mapstructure.Decode(c, &ttnC); err == nil {
				if strings.ToLower(ttnC.DeviceEUI) == strings.ToLower(rq.HardwareSerial) && strings.ToLower(ttnC.ApplicationID) == strings.ToLower(rq.AppID) {
					thingID = t.ID
				}
			}
		}
	}
	if thingID == "" {
		return c.Error(http.StatusNotFound, fmt.Errorf("Device %s on Application %s with ProjectID %s Not Found", rq.DevID, rq.AppID, projectID))
	}

	states := make(map[interface{}]interface{})
	if err := codec.NewDecoderBytes(rq.PayloadRaw, new(codec.CborHandle)).Decode(&states); err != nil {
		coreApp.Logger.WithFields(logrus.Fields{
			"component": "ttn service",
		}).Errorf("Incoming data from %s @ %s with pid: %s is not a valid cbor (%q) %s", rq.DevID, rq.AppID, projectID, rq.PayloadRaw, err)
		return c.Render(http.StatusOK, r.JSON(true))
	}

	for name, value := range states {
		state := types.State{
			Raw:     value,
			At:      rq.Metadata.Time,
			ThingID: thingID,
			Project: projectID,
			Asset:   fmt.Sprintf("%v", name), // convert anything to string (is there any better way?)
		}
		coreApp.Data(state)
	}

	return c.Render(http.StatusOK, r.JSON(true))
}
