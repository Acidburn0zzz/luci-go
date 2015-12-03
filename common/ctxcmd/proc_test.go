// Copyright 2015 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ctxcmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

const (
	envHelperTest = "GO_CTXCMD_HELPER_TEST=1"

	// waitForeverReady is text emitted by process when it has installed its
	// signal handler and begun waiting forever.
	waitForeverReady = "Waiting forever..."
)

func isHelperTest() bool {
	for _, v := range os.Environ() {
		if v == envHelperTest {
			return true
		}
	}
	return false
}

func helperCommand(t *testing.T, name string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], fmt.Sprintf("-test.run=%s", name))
	cmd.Env = []string{envHelperTest}
	return cmd
}

func TestCtxCmd(t *testing.T) {
	t.Parallel()

	Convey(`A cancellable context`, t, func() {
		c, cancelFunc := context.WithCancel(context.Background())

		Convey(`When running a bogus process`, func() {
			cc := CtxCmd{
				Cmd: exec.Command("##fake-does-not-exist##"),
			}

			Convey(`If the Context is already cancelled, the process will not run.`, func() {
				cancelFunc()
				So(cc.Start(c), ShouldEqual, context.Canceled)
			})
		})

		Convey(`When running a process that exits immediately`, func() {
			cc := CtxCmd{
				Cmd: helperCommand(t, "TestExitImmediately"),
			}

			Convey(`The process runs and exits successfully.`, func() {
				So(cc.Run(c), ShouldBeNil)
			})

			Convey(`Cancelling after process exit returns process' exit value.`, func() {
				cc.testProcFinishedC = make(chan struct{})
				So(cc.Start(c), ShouldBeNil)
				<-cc.testProcFinishedC

				// Cancel our Context.
				cancelFunc()

				// Make sure that we got a process return value.
				err := cc.Wait()
				So(err, ShouldBeNil)
				So(cc.ProcessError, ShouldBeNil)
			})
		})

		Convey(`When running a process that runs forever`, func() {
			cc := CtxCmd{
				Cmd: helperCommand(t, "TestWaitForever"),
			}

			Convey(`Cancelling the process causes it to exit with non-zero return code.`, func() {
				So(cc.Start(c), ShouldBeNil)
				cancelFunc()

				So(cc.Wait(), ShouldEqual, context.Canceled)

				So(cc.ProcessError, ShouldHaveSameTypeAs, (*exec.ExitError)(nil))

				exitErr := cc.ProcessError.(*exec.ExitError)
				So(exitErr.Sys().(syscall.WaitStatus).ExitStatus(), ShouldNotEqual, 0)
			})

			Convey(`Interrupting the process causes it to exit with return code 5 (see main()).`, func() {
				if runtime.GOOS == "windows" {
					// This test does not work on Windows, as processes cannot send
					// signals (e.g., os.Interrupt) to other processes.
					return
				}

				// Create a pipe to ensure that the process has started.
				rc, err := cc.Cmd.StdoutPipe()
				So(err, ShouldBeNil)
				defer rc.Close()

				cc.CancelSignal = os.Interrupt
				So(cc.Start(c), ShouldBeNil)

				// Read "Waiting forever..."
				buf := make([]byte, len(waitForeverReady))
				_, err = rc.Read(buf)
				So(err, ShouldBeNil)
				So(string(buf), ShouldEqual, waitForeverReady)

				// Process is ready, go ahead and cancel.
				cancelFunc()
				So(cc.Wait(), ShouldEqual, context.Canceled)

				So(cc.ProcessError, ShouldHaveSameTypeAs, (*exec.ExitError)(nil))

				exitErr := cc.ProcessError.(*exec.ExitError)
				So(exitErr.Sys().(syscall.WaitStatus).ExitStatus(), ShouldEqual, 5)
			})
		})
	})
}
