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
	"fmt"
	"testing"
	"time"

	"github.com/FANIoT/types"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/assert"
)

const tID = "el-thing" // ThingID
const aName = "memory" // Asset Name
const pName = "her"    // Project Name

func TestPipeline(t *testing.T) {
	a := New()
	a.Run()
	ts := time.Now()

	assert.NoError(t, a.Data(types.State{
		Raw:     18.20,
		At:      ts,
		Asset:   aName,
		ThingID: tID,
		Project: pName,
	}))
	a.Exit()

	var d types.State
	q := a.db.Collection(fmt.Sprintf("data.%s.%s", pName, tID)).FindOne(context.Background(), bson.NewDocument(
		bson.EC.SubDocument("at", bson.NewDocument(
			bson.EC.Time("$gte", ts),
		)),
		bson.EC.String("asset", aName),
	))
	assert.NoError(t, q.Decode(&d))

	assert.Equal(t, d.At.Unix(), ts.Unix())
	assert.Equal(t, 18.20, d.Value.Number)
}

func BenchmarkPipeline(b *testing.B) {
	a := New()
	a.Run()

	wait := make(chan struct{})
	a.cli.Subscribe(fmt.Sprintf("i1820/projects/%s/things/%s/assets/%s/state", pName, tID, aName), 0, func(client paho.Client, message paho.Message) {
		fmt.Println("Hello")
		wait <- struct{}{}
	})

	for i := 0; i < b.N; i++ {
		ts := time.Now()

		assert.NoError(b, a.Data(types.State{
			At:      ts,
			Raw:     18.20,
			Asset:   aName,
			ThingID: tID,
			Project: pName,
		}))

		<-wait
	}
}
