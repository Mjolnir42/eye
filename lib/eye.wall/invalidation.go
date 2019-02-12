/*-
 * Copyright © 2018, 1&1 Internet SE
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package wall // import "github.com/solnx/eye/lib/eye.wall"

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis"
	"github.com/mjolnir42/erebos"
)

// Invalidation implements the eyewall cache invalidation used by eye
type Invalidation struct {
	Config   *erebos.Config
	Registry map[string]*redis.Client
	sync.RWMutex
}

// NewInvalidation returns a new Invalidation
func NewInvalidation(conf *erebos.Config) (iv *Invalidation) {
	iv = &Invalidation{}
	iv.Config = conf
	iv.Registry = make(map[string]*redis.Client)
	return
}

// Register adds a new cache to the registry
func (iv *Invalidation) Register(regID, addr string, port, db int64) (err error) {
	iv.Lock()
	defer iv.Unlock()

	cl := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", addr, port),
		Password: iv.Config.Redis.Password,
		DB:       int(db),
	})
	if _, err = cl.Ping().Result(); err != nil {
		return
	}
	iv.Registry[regID] = cl
	return
}

// Unregister deletes a cache from the registry
func (iv *Invalidation) Unregister(regID string) {
	iv.Lock()
	iv.Registry[regID].Close()
	delete(iv.Registry, regID)
	iv.Unlock()
}

// CloseAll closes all active redis clients in iv.Registry
func (iv *Invalidation) CloseAll() {
	iv.Lock()
	for regID := range iv.Registry {
		iv.Registry[regID].Close()
		delete(iv.Registry, regID)
	}
	iv.Unlock()
}

// AsyncInvalidate removes lookupID from all registered caches. It calls
// Invalidate(lookupID) and handles the returned channels to avoid
// blocked resources.
func (iv *Invalidation) AsyncInvalidate(lookupID string) {
	go func() {
		done, errors := iv.Invalidate(lookupID)
		for {
			select {
			case <-errors:
			case <-done:
				break
			}
		}
	}()
	return
}

// Invalidate removes lookupID from all registered caches. Errors
// encountered are written to errors and done is closed once all caches
// have been updated.
// Both channels must be read.
func (iv *Invalidation) Invalidate(lookupID string) (done chan struct{}, errors chan error) {
	iv.RLock()

	go func(done chan struct{}, errors chan error) {
		defer iv.RUnlock()
		wg := sync.WaitGroup{}

		for cacheID := range iv.Registry {
			wg.Add(1)
			go func(c, l string) {
				defer wg.Done()

				errors <- iv.invalidateCache(c, l)
			}(cacheID, lookupID)
		}
		wg.Wait()
		close(done)
	}(done, errors)

	return
}

// invalidateCache implements removing lookupID from a single cache
// registered as cacheID
func (iv *Invalidation) invalidateCache(cacheID, lookupID string) error {
	iv.RLock()
	defer iv.RUnlock()

	// declare here to enable recursive definition
	var clear func(string, string) error

	clear = func(cache, key string) error {
		err := iv.Registry[cache].Watch(
			func(tx *redis.Tx) error {
				profiles, err := tx.HGetAll(key).Result()
				if err != nil && err != redis.Nil {
					return err
				}

				_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
					for profileID := range profiles {
						pipe.Del(profileID)
						pipe.HDel(key, profileID)
					}
					pipe.Del(key)
					return nil
				})
				return err
			},
			key,
		)

		if err == redis.TxFailedErr {
			return clear(cache, key)
		}
		return err
	}

	return clear(cacheID, lookupID)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
