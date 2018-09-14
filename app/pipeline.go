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

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"

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
		// Find thing in I1820/pm
		t, err := a.pm.ThingsShow(d.ThingID)
		if err != nil {
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Errorf("PM ThingsShow: %s", err)
		} else {
			d.Project = t.Project
			d.Model = t.Model
		}

		// if there isn't any project at this moment, it will ignore further processing and store raw format of data
		if d.Project != "" {
			// marshal data into json format
			b, err := json.Marshal(d)
			if err != nil {
				a.Logger.WithFields(logrus.Fields{
					"component": "link",
				}).Errorf("Marshal data error: %s", err)
			}

			// publish data with only its raw format
			a.cli.Publish(fmt.Sprintf("i1820/project/%s/raw", d.Project), 0, false, b)
			a.Logger.WithFields(logrus.Fields{
				"component": "link",
			}).Infof("Publish raw data: %s", d.Project)
		}

		a.decodeStream <- d
	}
}

// decodeStage decodes each data based on its things model
func (a *Application) decodeStage() {
	// This thread is mine
	runtime.LockOSThread()

	a.Logger.WithFields(logrus.Fields{
		"component": "link",
	}).Info("Decode pipeline stage")

	for d := range a.decodeStream {
		// Run decode only when data is coming from a thing with project and it needs decode
		if d.Project != "" && d.Data == nil {
			if d.Model != "generic" {
				m, ok := a.models[d.Model]
				if !ok {
					a.Logger.WithFields(logrus.Fields{
						"component": "link",
					}).Errorf("Model %s not found", d.Model)
				} else {
					d.Data = m.Decode(d.Raw)
				}

				// marshal data into json format
				b, err := json.Marshal(d)
				if err != nil {
					a.Logger.WithFields(logrus.Fields{
						"component": "link",
					}).Errorf("Marshal data error: %s", err)
				}

				// publish data with both raw and decoded formats
				a.cli.Publish(fmt.Sprintf("i1820/project/%s/data", d.Project), 0, false, b)
				a.Logger.WithFields(logrus.Fields{
					"component": "link",
				}).Infof("Publish parsed data: %s", d.Project)
			}
		}
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
		if _, err := a.db.Collection("data").InsertOne(context.Background(), d); err != nil {
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
