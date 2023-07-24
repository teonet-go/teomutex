// Copyright 2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// teomutex is a Golang package teomutex is Teonet Cloud Mutex baset on Google
// Cloud Storage. It can be used to serialize computations anywhere on the
// global internet.
//
// How to use it:
//
// - Create Google Cloud Storage bucket in which lock objects will be stored.
//     Use next command to create backet: `gsutil mb gs:mutex`. By default
//     the teomutex uses the "mutext" backet name. To use another backet name
//     set it in second parameter of the `teomutex.NewMutex` function.
//
// - In your application import the `github.com/teonet/teomutex` package,
//     and create new mutex:
// ```go
/*
	// Creates new Teonet Mutex object.
	m, err := teomutex.NewMutex("test/lock/some_object")
	if err != nil {
		// Process error
		return
	}
	defer m.Close()
*/
// ```
//
// - Use the `m.Lock` and `m.Unlock` functions to lock and unlock:
// ```go
/*
	// Lock mutex
	err = m.Lock()
	if err != nil {
		//  Process error
		return
	}

	// Do somthing in this protected area

	// Unlock mutex
	err = m.Unlock()
	if err != nil {
		// Process error
		return
	}
*/
// ```
package teomutex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

const (
	defaultBucket  = "mutex"
	defaultTimeout = 10 * time.Second
)

// Mutex object structure and methods receiver.
type Mutex struct {
	client  *storage.Client // Storage client
	bucket  string          // Bucket name
	object  string          // Object name
	timeout time.Duration   // Lock timeout
	w       io.Writer       // Log writer
}

// NewMutex creates new Teonet Mutex object.
//
// Parameters:
//
//	object - is the name of lock object
//	backet - is the name of backet where lock objects created, by default used the "mutex" backet
func NewMutex(object string, bucket ...string) (m *Mutex, err error) {

	// Creates new Mutex object
	m = new(Mutex)

	// Set backet and object name
	if len(bucket) > 0 {
		m.bucket = bucket[0]
	} else {
		m.bucket = defaultBucket
	}
	m.object = object

	// Set log writer
	m.w = os.NewFile(0, os.DevNull)

	// Creates storage client
	ctx := context.Background()
	m.client, err = storage.NewClient(ctx)
	if err != nil {
		err = fmt.Errorf("creates storage client error: %s", err)
		return
	}

	// Set default Lock timeout
	m.timeout = defaultTimeout

	return
}

// Close the Mutex object.
func (m Mutex) Close() error {
	return m.client.Close()
}

// SetLockTimeout sets lock timeout to avoid deadlock. The default timeout is
// set to 10 seconds.
func (m *Mutex) SetLockTimeout(timeout time.Duration) {
	m.timeout = timeout
}

// SetLogWriter sets log writer used in teomutex package functions.
func (m *Mutex) SetLogWriter(w io.Writer) {
	m.w = w
}

// Lock mutex
func (m Mutex) Lock() error {
	repeatAfter := 1 * time.Millisecond
	start := time.Now()
	for {
		if err := m.uploadObject(); err == nil {
			return nil
		} else {
			fmt.Fprintf(m.w, "%s\n", err)
		}
		timeout := m.timeout - time.Since(start)
		select {
		case <-time.After(repeatAfter):
			repeatAfter *= 2
			continue
		case <-time.After(timeout):
			return fmt.Errorf("lock timeout")
		}
	}
}

// Unock mutex
func (m Mutex) Unlock() error {
	return m.deleteObject()
}

// uploadObject uploads mutex object.
func (m Mutex) uploadObject() error {

	// Print start message
	fmt.Fprintf(m.w, "Uploading object %s started...\n", m.object)

	// Create bytes reader to upload
	r := bytes.NewReader([]byte("locked"))

	o := m.client.Bucket(m.bucket).Object(m.object)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	// For an object that does not yet exist, set the DoesNotExist precondition.
	o = o.If(storage.Conditions{DoesNotExist: true})
	// If the live object already exists in your bucket, set instead a
	// generation-match precondition using the live object's generation number.
	// attrs, err := o.Attrs(ctx)
	// if err != nil {
	//      return fmt.Errorf("object.Attrs: %w", err)
	// }
	// o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	// Upload an object with bytes.Reader
	wc := o.NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("writer.Close: %w", err)
	}

	// Print success result message
	fmt.Fprintf(m.w, "Blob %s uploaded.\n", m.object)
	return nil
}

// deleteObject deletess mutex object.
func (m Mutex) deleteObject() error {

	// Print start message
	fmt.Fprintf(m.w, "Deleting object %s started...\n", m.object)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	o := m.client.Bucket(m.bucket).Object(m.object)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to delete the file is aborted
	// if the object's generation number does not match your precondition.
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("object.Attrs: %w", err)
	}
	o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})

	if err := o.Delete(ctx); err != nil {
		return fmt.Errorf("object(%q).Delete: %w", m.object, err)
	}

	// Print success result message
	fmt.Fprintf(m.w, "Blob %s deleted.\n", m.object)
	return nil
}
