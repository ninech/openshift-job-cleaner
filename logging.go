package main

import (
	"errors"
	"github.com/evalphobia/logrus_sentry"
	log "github.com/sirupsen/logrus"
	"os"
)

// continueOnError will print an error if passed in and return true or false based on error state,
// mainly used to continue inside loops
func continueOnError(err error) bool {
	if err != nil {
		log.Error(err)
		return true
	}
	return false
}

// panicOnError will send an error to logrus and panic if an error is passed in
func panicOnError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// infoAndExit prints an info level message and then exits if exitCode >=0
func infoAndExit(msg string, exitCode int) {
	log.Info(msg)
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}

// initLogging sets up sentry
func initLogging(sentryDsn string) error {
	if sentryDsn == "" {
		return errors.New("SENTRY_DSN cannot be blank")
	}
	hook, err := logrus_sentry.NewSentryHook(sentryDsn, []log.Level{
		log.PanicLevel,
		log.FatalLevel,
		log.ErrorLevel,
	})
	if err != nil {
		return err
	}
	hook.StacktraceConfiguration.Enable = true
	log.AddHook(hook)
	return nil
}
