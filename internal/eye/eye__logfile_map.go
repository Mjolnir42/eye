/*-
 * Copyright (c) 2017, Jörg Pernfuß
 * All rights reserved.
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package eye // import "github.com/solnx/eye/internal/eye"

import (
	"sync"

	"github.com/client9/reopen"
)

// LogHandleMap is a concurrent map that is used to look up
// filehandles of active logfiles
type LogHandleMap struct {
	hmap map[string]*reopen.FileWriter
	sync.RWMutex
}

// Initialize the LogHandleMap
func (l *LogHandleMap) Init() {
	l.hmap = make(map[string]*reopen.FileWriter)
}

// Add registers a new filehandle
func (l *LogHandleMap) Add(key string, fh *reopen.FileWriter) {
	l.Lock()
	defer l.Unlock()
	l.hmap[key] = fh
}

// Get retrieves a filehandle
func (l *LogHandleMap) Get(key string) *reopen.FileWriter {
	l.RLock()
	defer l.RUnlock()
	return l.hmap[key]
}

// Del removes a filehandle
func (l *LogHandleMap) Del(key string) {
	l.Lock()
	defer l.Unlock()
	delete(l.hmap, key)
}

// Range returns the keys of all registered filehandles
func (l *LogHandleMap) Range() chan string {
	l.RLock()
	resChan := make(chan string, len(l.hmap))
	go func(out chan string) {
		for name := range l.hmap {
			out <- name
		}
		close(out)
		l.RUnlock()
	}(resChan)
	return resChan
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
