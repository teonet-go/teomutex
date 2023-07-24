// Copyright 2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package teomutex

import (
	"fmt"
	"os"
	"testing"
	"time"
)

var printLine = func() {
	fmt.Println("---")
}

func TestLockUnlock(t *testing.T) {

	// Creates new Teonet Mutex object
	m, err := NewMutex("test/lock/some_object")
	if err != nil {
		t.Error(err)
		return
	}
	defer m.Close()
	m.SetLockTimeout(10 * time.Millisecond)
	m.SetLogWriter(os.Stdout)

	// Lock mutex
	err = m.Lock()
	if err != nil {
		t.Error(err)
		return
	}
	printLine()

	// Lock mutex again (err after timeout)
	err = m.Lock()
	if err == nil {
		t.Error("lock error: locks already locked mutex without error")
		return
	}
	t.Log(err)
	printLine()

	// Unlock mutex
	err = m.Unlock()
	if err != nil {
		t.Error(err)
		return
	}
	printLine()

	// Unlock mutex again (err)
	err = m.Unlock()
	if err == nil {
		t.Error("unlock error: unlocks doesn't locked mutex without error")
		return
	}
	t.Log(err)
	printLine()
}

func TestWaitUnlock(t *testing.T) {

	// Creates new Teonet Mutex object
	m, err := NewMutex("test/lock/some_object")
	if err != nil {
		t.Error(err)
		return
	}
	defer m.Close()
	// Use default timeout 10 sec
	// m.SetLockTimeout(10 * time.Second)
	m.SetLogWriter(os.Stdout)

	// Lock mutex
	err = m.Lock()
	if err != nil {
		t.Error(err)
		return
	}
	printLine()

	// Unlock mutex after 1 sec
	time.AfterFunc(1*time.Second, func() {
		err = m.Unlock()
		if err != nil {
			t.Error(err)
			return
		}
		printLine()
	})

	// Lock mutex again
	err = m.Lock()
	if err != nil {
		t.Error(err)
		return
	}
	printLine()

	// Unlock mutex
	err = m.Unlock()
	if err != nil {
		t.Error(err)
		return
	}
	printLine()
}
