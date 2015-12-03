// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// +build darwin dragonfly freebsd linux netbsd openbsd

package streamserver

import (
	"net"
	"os"

	log "github.com/luci/luci-go/common/logging"
	"golang.org/x/net/context"
)

// NewNamedPipeServer instantiates a new POSIX named pipe server instance.
func NewNamedPipeServer(ctx context.Context, path string) StreamServer {
	ctx = log.SetField(ctx, "namedPipePath", path)
	return &listenerStreamServer{
		Context: ctx,
		gen: func() (net.Listener, error) {
			log.Infof(ctx, "Creating POSIX server socket Listener.")

			// Cleanup any previous named pipe. We don't bother checking for the file
			// first since the remove is atomic. We also ignore any error here, since
			// it's probably related to the file not being found.
			//
			// If there was an actual error removing the file, we'll catch it shortly
			// when we try to create it.
			os.Remove(path)

			// Create a UNIX listener
			l, err := net.Listen("unix", path)
			if err != nil {
				return nil, err
			}

			ul := selfCleaningUNIXListener{
				Context:  ctx,
				Listener: l,
				path:     path,
			}
			return &ul, nil
		},
	}
}

// Wrapper around the "unix"-type Listener that cleans up the named pipe on
// creation and 'Close()'
type selfCleaningUNIXListener struct {
	context.Context
	net.Listener

	path string
}

func (l *selfCleaningUNIXListener) Close() error {
	if err := l.Listener.Close(); err != nil {
		return err
	}

	if err := os.Remove(l.path); err != nil {
		log.Fields{
			log.ErrorKey: err,
		}.Warningf(l, "Failed to remove named pipe file on Close().")
	}
	return nil
}
