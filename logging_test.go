package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestInitLoggingFail(t *testing.T) {
	err := initLogging("")
	if err == nil {
		t.Error("blank sentry dns values should return an error")
	}
}

func TestInitLoggingMalformedDSN(t *testing.T) {
	err := initLogging("https://sentry.io/<project>")
	if err == nil {
		t.Error("malformed sentry dns values should return an error")
	}
}

func TestInitLogging(t *testing.T) {
	err := initLogging("https://<key>:<secret>@sentry.io/<project>")
	if err != nil {
		t.Error(err)
	}
}

func TestContinueOnError(t *testing.T) {
	if continueOnError(nil) {
		t.Error("should not say there was an error when there was not")
	}
	if !continueOnError(errors.New("test error")) {
		t.Error("should not say there was no error when there was")
	}
}

func TestFailOnErrorNoError(t *testing.T) {
	//if this fails everything will blow up, so we dont need any assertion
	panicOnError(nil)
}

//because of how we have to test of exiting errors code coverage will never respect that we did the tests below here.

func TestFailOnError(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		panicOnError(errors.New("test error"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFailOnError")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestInfoAndExit(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		infoAndExit("exiting with code", 5)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestInfoAndExit")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want non zero exit code", err)
}

func TestInfoAndExitClean(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		infoAndExit("exiting with zero", 0)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestInfoAndExitClean")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		t.Fatalf("process ran with err %v, want zero exit code", err)
	}

}

func TestInfoAndExitShortCircuit(t *testing.T) {
	infoAndExit("will not exit", -1)
}
