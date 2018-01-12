package main

import (
	"fmt"
	"time"
)

// discardSuccessTimestamps takes a slice of job items and removes:
// 		all the elements earlier than earliest time
//		any item which has a year of 0001 indicating an invalid status from the request (ie job still in progress)
func discardSuccessTimestamps(items []jobItem, earliest time.Time) []jobItem {
	older, _ := sort(items, func(item jobItem) bool {
		if time.Time(item.Status.CompletionTime).Year() == 0001 {
			return false
		}
		return time.Time(item.Status.CompletionTime).Before(earliest)
	})
	return older
}

// discardFailureTimestamps takes a slice of job items and removes all the elements earlier than earliest
func discardFailureTimestamps(items []jobItem, earliest time.Time) []jobItem {
	older, _ := sort(items, func(item jobItem) bool {
		if len(item.Status.Conditions) == 0 {
			return false
		}
		return item.Status.Failed == 1 && time.Time(item.Status.Conditions[0].LastTransitionTime).Before(earliest)
	})
	return older
}

// sort wraps a user supplied function and returns two slices depending on the success or failure state of the supplied function for each
func sort(vs []jobItem, f func(item jobItem) bool) ([]jobItem, []jobItem) {
	sortsTrue := make([]jobItem, 0)
	sortsFalse := make([]jobItem, 0)
	for _, v := range vs {
		if f(v) {
			sortsTrue = append(sortsTrue, v)
		} else {
			sortsFalse = append(sortsFalse, v)
		}
	}
	return sortsTrue, sortsFalse
}

// buildMaxAgeDuration is a simple helper to form a basic int in the app config file to a time string we can use as a duration
// userVal is in hours
func buildMaxAgeDuration(userVal int) (time.Time, error) {
	retentionActual := fmt.Sprintf("-%dm", userVal)
	rDur, err := time.ParseDuration(retentionActual)
	if err != nil {
		// This should only happen if the input is a negative value
		return time.Time{}, err
	}
	return time.Now().Add(rDur), nil
}
