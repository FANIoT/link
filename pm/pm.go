/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 20-09-2018
 * |
 * | File Name:     pm.go
 * +===============================================
 */

package pm

import (
	"context"
	"fmt"
	"log"

	"github.com/I1820/types"
	"github.com/gobuffalo/envy"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	cache "github.com/patrickmn/go-cache"

	"github.com/mongodb/mongo-go-driver/bson"
)

var db *mgo.Database

func init() {
	// initiate mongo connection
	url := envy.Get("DB_URL", "mongodb://172.18.0.1:27017")
	client, err := mgo.NewClient(url)
	if err != nil {
		log.Fatalf("DB new client error: %s", err)
	}
	if err := client.Connect(context.Background()); err != nil {
		log.Fatalf("DB connection error: %s", err)
	}
	db = client.Database("i1820")

}

// ThingByID finds thing by its id in pm component database.
func ThingByID(ctx context.Context, id string) (types.Thing, error) {
	// check cache in the first place
	if th, found := c.Get(id); found {
		return th.(types.Thing), nil
	}

	var t types.Thing
	// find things by its id (please note that it must be activated)
	dr := db.Collection("things").FindOne(ctx, bson.NewDocument(
		bson.EC.Boolean("status", true),
		bson.EC.String("_id", id),
	))
	if err := dr.Decode(&t); err != nil {
		if err == mgo.ErrNoDocuments {
			return t, fmt.Errorf("Thing %s not found", id)
		}
		return t, err
	}

	// Set the value of the key thing_id to thing, with the default expiration time
	c.Set(id, t, cache.DefaultExpiration)

	return t, nil
}
