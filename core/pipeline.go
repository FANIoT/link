/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 02-08-2018
 * |
 * | File Name:     pipeline.go
 * +===============================================
 */

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/FANIoT/link/pm"
	"github.com/FANIoT/types"
	"github.com/sirupsen/logrus"
)

// projectStage finds project for each data based on its thing identification.
func (a *Application) projectStage() {
	// This thread is mine
	runtime.LockOSThread()

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Project pipeline stage")

	for d := range a.projectStream {
		// retrieve project when it is needed
		if d.Project == "" {
			t, err := pm.ThingByID(context.Background(), d.ThingID)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"component": "link",
					"asset":     d.Asset,
					"thingid":   d.ThingID,
				}).Errorf("Project find error: %s", err)
				continue
			}
			d.Project = t.Project
		}

		a.decodeStream <- d
	}

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Project pipeline stage is going out")
	select {
	case a.projectCloseChan <- struct{}{}:
		close(a.decodeStream)
	default:
	}
}

// decodeStage decodes each data and fills value section.
// as you see there is no specific decode happens here so models
// must do they job somewhere else
func (a *Application) decodeStage() {
	// This thread is mine
	runtime.LockOSThread()

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Decode pipeline stage")

	for d := range a.decodeStream {
		switch v := d.Raw.(type) { // find type of raw value
		case string: // string
			d.Value.String = v
		case bool: // boolean
			d.Value.Boolean = v
		case float64: // number
			d.Value.Number = v
		case float32:
			d.Value.Number = float64(v)
		case int:
			d.Value.Number = float64(v)
		case int8:
			d.Value.Number = float64(v)
		case int16:
			d.Value.Number = float64(v)
		case int32:
			d.Value.Number = float64(v)
		case int64:
			d.Value.Number = float64(v)
		case uint:
			d.Value.Number = float64(v)
		case uint8:
			d.Value.Number = float64(v)
		case uint16:
			d.Value.Number = float64(v)
		case uint32:
			d.Value.Number = float64(v)
		case uint64:
			d.Value.Number = float64(v)
		case interface{}: // object
			d.Value.Object = d.Raw
		case []interface{}: // array
			d.Value.Array = d.Raw.([]interface{})
		}

		go func(d types.State) {
			// marshal data into json
			b, err := json.Marshal(d)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"component": "link",
					"asset":     d.Asset,
					"thingid":   d.ThingID,
				}).Errorf("Marshal data error: %s", err)
			}

			// publish data with both raw and typed formats
			// i1820/projects/{project_id}/things/{thing_id}/assets/{asset_name}/state
			a.cli.Publish(fmt.Sprintf("i1820/projects/%s/things/%s/assets/%s/state", d.Project, d.ThingID, d.Asset), 0, false, b)
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
				"asset":     d.Asset,
				"thingid":   d.ThingID,
			}).Infof("Publish decoded data: %s", d.Project)
		}(*d)
		a.Logger.WithFields(logrus.Fields{
			"component": "link",
			"asset":     d.Asset,
			"thingid":   d.ThingID,
		}).Infof("Decode with value: %+v", d.Value)

		a.insertStream <- d
	}

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Decode pipeline stage is going")
	select {
	case a.decodeCloseChan <- struct{}{}:
		close(a.insertStream)
	default:
	}
}

// insertStage inserts each data to database
func (a *Application) insertStage() {
	// This thread is mine
	runtime.LockOSThread()

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Insert pipeline stage")

	for d := range a.insertStream {
		if _, err := a.db.Collection(fmt.Sprintf("data.%s.%s", d.Project, d.ThingID)).InsertOne(context.Background(), *d); err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
				"asset":     d.Asset,
				"thingid":   d.ThingID,
			}).Errorf("Mongo Insert: %s", err)
		} else {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
				"asset":     d.Asset,
				"thingid":   d.ThingID,
			}).Infof("Insert into database with value: %+v", d.Value)
		}
	}

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Insert pipeline stage is going")
	a.insertCloseCounter.Done()
}
