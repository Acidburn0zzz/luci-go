// Copyright 2015 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package errors

import (
	"fmt"
)

// MultiError is a simple `error` implementation which represents multiple
// `error` objects in one.
type MultiError []error

func (m MultiError) Error() string {
	n, e := m.Summary()
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return e.Error()
	case 2:
		return e.Error() + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", e, n-1)
}

// Summary gets the total count of non-nil errors and returns the first one.
func (m MultiError) Summary() (n int, first error) {
	for _, e := range m {
		if e != nil {
			if n == 0 {
				first = e
			}
			n++
		}
	}
	return
}

// First returns the first non-nil error.
func (m MultiError) First() error {
	for _, e := range m {
		if e != nil {
			return e
		}
	}
	return nil
}

// StackContext implements StackContexter.
func (m MultiError) StackContext() StackContext {
	n, _ := m.Summary()

	return StackContext{
		InternalReason: "MultiError %(non-nil)d/%(total)d: following first non-nil error.",
		Data: Data{
			"non-nil": {Value: n},
			"total":   {Value: len(m)},
		},
	}
}

// NewMultiError create new multi error from given errors.
//
// Can be used to workaround 'go vet' confusion "composite literal uses unkeyed
// fields" or if you do not want to remember that MultiError is in fact []error.
func NewMultiError(errors ...error) MultiError {
	return errors
}

// SingleError provides a simple way to uwrap a MultiError if you know that it
// could only ever contain one element.
//
// If err is a MultiError, return its first element. Otherwise, return err.
func SingleError(err error) error {
	if me, ok := err.(MultiError); ok {
		if len(me) == 0 {
			return nil
		}
		return me[0]
	}
	return err
}
