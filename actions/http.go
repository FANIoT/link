/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 09-10-2018
 * |
 * | File Name:     http.go
 * +===============================================
 */

package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FANIoT/link/pm"
	"github.com/FANIoT/types"
	"github.com/gobuffalo/buffalo"
	"github.com/sirupsen/logrus"
	"github.com/ugorji/go/codec"
)

// HTTPAuthorize authorizes each state request on with its access token and thingid.
// Each thing have unique identification and array of access tokens, this function
// read thing identification from url and access token from `Authorization` header so
// consider that this function must be bind in followin path
// /things/{thing_id}
func HTTPAuthorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		authString := c.Request().Header.Get("Authorization")
		thingID := c.Param("thing_id")

		t, err := pm.ThingByID(c, thingID)
		if err != nil {
			return c.Error(http.StatusInternalServerError, err)
		}
		c.Set("project_id", t.Project)
		c.Set("thing_id", thingID)

		for _, token := range t.Tokens {
			if token == authString {
				return next(c)
			}
		}

		return c.Error(http.StatusUnauthorized, fmt.Errorf("unathorized access token"))
	}
}

// HTTPHandler handles state request that are coming from devices.
// it passes them into link pipeline.
func HTTPHandler(c buffalo.Context) error {
	thingID := c.Value("thing_id").(string)
	projectID := c.Value("project_id").(string)

	var h codec.Handle
	ct := c.Request().Header.Get("Content-Type")
	switch ct {
	case "application/json":
		h = new(codec.JsonHandle)
	case "application/cbor":
		h = new(codec.CborHandle)
	default:
		return c.Error(http.StatusBadRequest, fmt.Errorf("unsupported content type"))
	}

	states := make(map[interface{}]interface{})

	if err := codec.NewDecoder(c.Request().Body, h).Decode(&states); err != nil {
		coreApp.Logger.WithFields(logrus.Fields{
			"component": "http service",
		}).Errorf("Incoming data from %s with pid: %s is not a valid %s: %s", thingID, projectID, h.Name(), err)

		return c.Render(http.StatusOK, r.JSON(true))
	}

	for name, value := range states {
		state := types.State{
			Raw:     value,
			At:      time.Now(),
			ThingID: thingID,
			Project: projectID,
			Asset:   fmt.Sprintf("%v", name), // convert anything to string (is there any better way?)
		}
		coreApp.Data(state)
	}

	return c.Render(http.StatusOK, r.JSON(true))
}
