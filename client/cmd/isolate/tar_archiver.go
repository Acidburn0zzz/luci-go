// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha1"
	"hash"
	"io"
	"os"

	"github.com/luci/luci-go/common/isolated"
)

// osOpen wraps os.Open to allow faking out during tests.
var osOpen = func(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Tar represents a completed tar archive.
type Tar struct {
	Content []byte
	Digest  isolated.HexDigest

	// FileCount is the number of individual files inside the resultant tar.
	FileCount int
	// FileSize is the size (in bytes) of the individual files. It will be
	// smaller than the size of the resultant tar given by len(Content).
	FileSize int64

	releasec chan<- []byte // Give back the bytes on release.
}

// Release frees the resources being held by Tar. You must call Release once
// the given Tar is no longer required. Release invalidates Content.
func (t *Tar) Release() {
	buf := t.Content[:0]
	t.Content = nil
	t.releasec <- buf
	t.releasec = nil // Sanity check that we can't double-release.
}

// tarWriter wraps a single tar.Writer, but keeps track of the digest and total
// size as it builds the tar.
type tarWriter struct {
	buf  *bytes.Buffer
	tw   *tar.Writer
	h    hash.Hash
	n    int
	size int64
}

// newTarWriter returns a new writer. It will use b as the underyling storage
// for buffering.
func newTarWriter(b []byte) *tarWriter {
	h := sha1.New()
	buf := bytes.NewBuffer(b)
	return &tarWriter{
		buf: buf,
		h:   h,
		tw:  tar.NewWriter(io.MultiWriter(buf, h)),
	}
}

func (w *tarWriter) Write(item *Item) error {
	file, err := osOpen(item.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := w.tw.WriteHeader(&tar.Header{
		Name:     item.RelPath,
		Mode:     int64(item.Mode),
		Typeflag: tar.TypeReg,
		Size:     item.Size,
	}); err != nil {
		return err
	}
	if _, err := io.Copy(w.tw, file); err != nil {
		return err
	}

	w.n++
	w.size += item.Size
	return nil
}

func (w *tarWriter) Len() int {
	return w.buf.Len()
}

// FlushTar closes the writer and returns the resultant Tar. It is given
// a channel on which the underlying buffer should be sent upon Release.
// After the call to FlushTar, the tarWriter must no longer be used.
func (w *tarWriter) FlushTar(buffc chan<- []byte) *Tar {
	w.tw.Close()
	w.tw = nil // Sanity check that the tarWriter is not used again.
	return &Tar{
		Content:   w.buf.Bytes(),
		Digest:    isolated.Sum(w.h),
		FileCount: w.n,
		FileSize:  w.size,
		releasec:  buffc,
	}
}

// TarArchiver combines small Items into larger tar-archived items, represented
// by the Tar struct.
type TarArchiver struct {
	// buffc contains bytes slices ready for use in creating tarWriters. This
	// channel controls the number of tars that may be in flight concurrently.
	buffc chan []byte
	// w is the current tarWriter, if any.
	w *tarWriter
	// callback is invoked each time an archive is ready.
	callback func(*Tar)
}

// NewTarAchiver returns a new TarArchiver. The provided callback is invoked
// everytime a combined tar-archive item is ready.
func NewTarAchiver(callback func(*Tar)) *TarArchiver {
	// bufSize is the number of archivers that we will hold in memory
	// at any given time.
	const bufSize = 20

	ta := &TarArchiver{
		buffc:    make(chan []byte, bufSize),
		callback: callback,
	}
	for i := 0; i < bufSize; i++ {
		ta.buffc <- nil
	}
	return ta
}

// AddItem reads the given item, adding it to a tar archive.
// AddItem will block if there are no currently-available tar writers.
func (ta *TarArchiver) AddItem(item *Item) error {
	if ta.w == nil {
		ta.w = newTarWriter(<-ta.buffc)
	}
	if err := ta.w.Write(item); err != nil {
		return err
	}

	// If the archive is large enough, flush it.
	if ta.w.Len() >= archiveSizeTrigger {
		ta.Flush()
	}
	return nil
}

// Flush emits any currently enqueued tar contents, if any.
func (ta *TarArchiver) Flush() {
	w := ta.w
	if w == nil || w.Len() == 0 {
		return
	}
	ta.w = nil

	// Flush out the tarWriter to get the resultant Tar.
	tar := w.FlushTar(ta.buffc)
	ta.callback(tar)
}