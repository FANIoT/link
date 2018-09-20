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

	"github.com/I1820/link/pm"
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
		t, err := pm.ThingByID(context.Background(), d.ThingID)
		if err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Errorf("Project find error: %s", err)
			continue
		}
		d.Project = t.Project

		a.decodeStream <- d
	}
}

// decodeStage decodes each data and fills value section.
func (a *Application) decodeStage() {
	// This thread is mine
	runtime.LockOSThread()

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Decode pipeline stage")

	for d := range a.decodeStream {
		switch d.Raw.(type) {
		case string:
			d.Value.String = d.Raw.(string)
		case bool:
			d.Value.Boolean = d.Raw.(bool)
		case float64:
			d.Value.Number = d.Raw.(float64)
		case interface{}:
			d.Value.Object = d.Raw
		case []interface{}:
			d.Value.Array = d.Raw.([]interface{})
		}

		// marshal data into json
		b, err := json.Marshal(d)
		if err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Errorf("Marshal data error: %s", err)
		}

		// publish data with both raw and typed formats
		// i1820/projects/{project_id}/things/{thing_id}/assets/{asset_name}/state
		a.cli.Publish(fmt.Sprintf("i1820/projects/%s/things/%s/assets/%s/state", d.Project, d.ThingID, d.Asset), 0, false, b)
		a.Logger.WithFields(logrus.Fields{
			"component": "link",
		}).Infof("Publish decoded data: %s", d.Project)

		a.insertStream <- d
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
		if _, err := a.db.Collection(fmt.Sprintf("data.%s.%s", d.Project, d.ThingID)).InsertOne(context.Background(), d); err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Errorf("Mongo Insert: %s", err)
		} else {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Infof("Insert into database: %#v", d)
		}
	}
}
