/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 03-08-2018
 * |
 * | File Name:     app_test.go
 * +===============================================
 */

package core

import (
	"context"
	"testing"
	"time"

	"github.com/I1820/types"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/assert"
)

const tID = "el-thing"
const aName = "memory"

func TestPipeline(t *testing.T) {
	a := New()
	a.Run()
	ts := time.Now()

	a.Data(types.State{
		At:      ts,
		Asset:   aName,
		ThingID: tID,
		Project: "nothing",
	})
	// wait until data traverse pipeline
	time.Sleep(1 * time.Second)

	var d types.State
	q := a.db.Collection("data").FindOne(context.Background(), bson.NewDocument(
		bson.EC.SubDocument("timestamp", bson.NewDocument(
			bson.EC.Time("$gte", ts),
		)),
		bson.EC.String("thingid", "el-thing"),
	))
	assert.NoError(t, q.Decode(&d))

	assert.Equal(t, d.At.Unix(), ts.Unix())
}

func BenchmarkPipeline(b *testing.B) {
	a := New()
	a.Run()

	wait := make(chan struct{})
	a.cli.Subscribe("i1820/project/her/raw", 0, func(client paho.Client, message paho.Message) {
		wait <- struct{}{}
	})

	for i := 0; i < b.N; i++ {
		ts := time.Now()

		a.Data(types.State{
			At:      ts,
			Asset:   aName,
			ThingID: tID,
			Project: "nothing",
		})

		<-wait
	}
}
