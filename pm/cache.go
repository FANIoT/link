/*
 *
 * In The Name of God
 *
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 20-09-2018
 * |
 * | File Name:     cache.go
 * +===============================================
 */

package pm

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

// cache caches things that are fetched from database for faster access.
var c *cache.Cache

func init() {
	// initiate cache
	c = cache.New(5*time.Minute, 10*time.Minute)
}
