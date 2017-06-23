// Copyright 2016 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package errors

import (
	"fmt"
)

func someProcessingFunction(val int) error {
	if val == 1 {
		// New and Reason automatically include stack information.
		return Reason("bad number: %(val)d").D("val", val).Err()
	}
	if err := someProcessingFunction(val - 1); err != nil {
		// correctly handles recursion
		return Annotate(err).D("val", val).Err()
	}
	return nil
}

func someLibFunc(vals ...int) error {
	for i, v := range vals {
		if err := someProcessingFunction(v); err != nil {
			return (Annotate(err).Reason("processing %(val)d").
				D("val", v).
				D("i", i).
				D("secret", "value", "%s").Err())
		}
	}
	return nil
}

type MiscWrappedError struct{ error }

func (e *MiscWrappedError) Error() string     { return fmt.Sprintf("super wrapper(%s)", e.error.Error()) }
func (e *MiscWrappedError) InnerError() error { return e.error }

func errorWrapper(err error) error {
	if err != nil {
		err = &MiscWrappedError{err}
	}
	return err
}

func someIntermediateFunc(vals ...int) error {
	errch := make(chan error)
	go func() {
		defer close(errch)
		errch <- Annotate(errorWrapper(someLibFunc(vals...))).Reason("could not process").Err()
	}()
	me := MultiError(nil)
	for err := range errch {
		if err != nil {
			me = append(me, err)
		}
	}
	if me != nil {
		return Annotate(me).Reason("while processing %(vals)v").D("vals", vals).Err()
	}
	return nil
}

func ExampleAnnotate() {
	if err := someIntermediateFunc(3); err != nil {
		err = Annotate(err).Reason("top level").Err()
		fmt.Println("Public-facing error:\n ", err)
		fmt.Println("\nfull error:")
		for _, l := range FixForTest(RenderStack(err).ToLines("runtime", "_test")) {
			fmt.Println(l)
		}
	}

	// Output:
	// Public-facing error:
	//   top level: while processing [3]: could not process: super wrapper(processing 3: bad number: 1)
	//
	// full error:
	// GOROUTINE LINE
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:14 - errors.someProcessingFunction()
	//   reason: "bad number: 1"
	//   "val" = 1
	//
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:16 - errors.someProcessingFunction()
	//   "val" = 2
	//
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:16 - errors.someProcessingFunction()
	//   "val" = 3
	//
	// From frame 2 to 3, the following wrappers were found:
	//   unknown wrapper *errors.MiscWrappedError
	//
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:25 - errors.someLibFunc()
	//   reason: "processing 3"
	//   "i" = 0
	//   "secret" = value
	//   "val" = 3
	//
	// From frame 3 to 4, the following wrappers were found:
	//   MultiError 1/1: following first non-nil error.
	//     "non-nil" = 1
	//     "total" = 1
	//
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:51 - errors.someIntermediateFunc.func1()
	//   reason: "could not process"
	//
	// ... skipped SOME frames in pkg "runtime"...
	//
	// GOROUTINE LINE
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:60 - errors.someIntermediateFunc()
	//   reason: "while processing [3]"
	//   "vals" = []int{3}
	//
	// #? github.com/luci/luci-go/common/errors/annotate_example_test.go:66 - errors.ExampleAnnotate()
	//   reason: "top level"
	//
	// #? testing/example.go:XXX - testing.runExample()
	// #? testing/example.go:XXX - testing.runExamples()
	// #? testing/testing.go:XXX - testing.(*M).Run()
	// ... skipped SOME frames in pkg "_test"...
	// ... skipped SOME frames in pkg "runtime"...
}
