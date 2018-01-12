package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestBuildRetentionDuration(t *testing.T) {

	rand.Seed(time.Now().Unix())

	for i := 0; i < 100; i++ {

		base := rand.Int()
		testTime := fmt.Sprintf("-%dh", base)

		minBound, _ := time.ParseDuration("-1s")
		maxBound, _ := time.ParseDuration("1s")
		dur, _ := time.ParseDuration(testTime)
		expectedMin := time.Now().Add(dur).Add(minBound)
		expectedMax := time.Now().Add(dur).Add(maxBound)

		retention, err := buildRetentionDuration(0)
		if err != nil {
			t.Fatal("should not produce error for valid integer")
		}
		if retention.After(time.Now()) {
			t.Fatal("retention can never be in the future")
		}
		if retention.Before(expectedMin) || retention.After(expectedMax) {
			t.Fatal("value before expected min", retention, "Expected Min:", expectedMin)
		}
		if retention.After(expectedMax) {
			t.Fatal("value after expected max time:", retention, "Expected Max:", expectedMax)
		}
	}
}

func TestFutureRetentionTime(t *testing.T) {
	_, err := buildRetentionDuration(-1)
	if err == nil {
		t.Fatal("retention time error should be created if time string is invalid")
	}
}

func TestSort(t *testing.T) {

	testData := make([]jobItem, 4)
	testData[0].Status.Succeeded = 0
	testData[1].Status.Succeeded = 1
	testData[2].Status.Succeeded = 23
	testData[3].Status.Succeeded = 666
	//func(item jobItem) bool
	success, failure := sort(testData, func(item jobItem) bool {
		return item.Status.Succeeded > 0
	})
	if len(success) != 3 {
		t.Fatal("should have 3 successes", success, testData)
	}
	if len(failure) != 1 {
		t.Fatal("should have 1 failures", failure, testData)
	}
}

func TestDiscardSuccessTimestamps(t *testing.T) {

	testData := make([]jobItem, 5)
	//processor needs time formats to be the same as the openshift output
	// 2017-12-01T14:00:17Z
	testData[0].Status.CompletionTime = ocTime(time.Now().Add(timestampHelper(1)))
	testData[0].Status.Succeeded = 0
	testData[1].Status.CompletionTime = ocTime(time.Now().Add(timestampHelper(16)))
	testData[1].Status.Succeeded = 1
	testData[2].Status.CompletionTime = ocTime(time.Now().Add(timestampHelper(5)))
	testData[2].Status.Succeeded = 2
	testData[3].Status.CompletionTime = ocTime(time.Now().Add(timestampHelper(25)))
	testData[3].Status.Succeeded = 3
	testData[4].Status.CompletionTime = ocTime{}
	testData[4].Status.Succeeded = 3

	retentionTime, _ := buildRetentionDuration(1)
	finaloutput := discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 4, t)

	retentionTime, _ = buildRetentionDuration(15)
	finaloutput = discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 2, t)
	if finaloutput[0].Status.Succeeded != 1 || finaloutput[1].Status.Succeeded != 3 {
		t.Fatal("returned the wrong items", finaloutput)
	}

	retentionTime, _ = buildRetentionDuration(16)
	finaloutput = discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 2, t)

	retentionTime, _ = buildRetentionDuration(17)
	finaloutput = discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 1, t)

	retentionTime, _ = buildRetentionDuration(25)
	finaloutput = discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 1, t)

	retentionTime, _ = buildRetentionDuration(26)
	finaloutput = discardSuccessTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 0, t)

}

func TestDiscardFailureTimestampsUninitializedCondition(t *testing.T) {
	testData := make([]jobItem, 1)
	testData[0].Status.Failed = 1
	retentionTime, _ := buildRetentionDuration(1)
	finaloutput := discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 0, t)
}

func TestDiscardFailureTimestamps(t *testing.T) {

	testData := make([]jobItem, 4)
	//processor needs time formats to be the same as the openshift output
	// 2017-12-01T14:00:17Z
	tNow := time.Now()
	testData[0].Status.Conditions = make([]condition, 1)
	testData[0].Status.Conditions[0].LastTransitionTime = ocTime(tNow.Add(timestampHelper(1)))
	testData[0].Status.Failed = 0
	testData[1].Status.Conditions = make([]condition, 1)
	testData[1].Status.Conditions[0].LastTransitionTime = ocTime(tNow.Add(timestampHelper(16)))
	testData[1].Status.Failed = 1
	testData[2].Status.Conditions = make([]condition, 1)
	testData[2].Status.Conditions[0].LastTransitionTime = ocTime(tNow.Add(timestampHelper(5)))
	testData[2].Status.Failed = 1
	testData[3].Status.Conditions = make([]condition, 1)
	testData[3].Status.Conditions[0].LastTransitionTime = ocTime(tNow.Add(timestampHelper(25)))
	testData[3].Status.Failed = 1

	retentionTime, _ := buildRetentionDuration(1)
	finaloutput := discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 3, t)

	retentionTime, _ = buildRetentionDuration(15)
	finaloutput = discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 2, t)
	if finaloutput[0].Status.Failed != 1 || finaloutput[1].Status.Failed != 1 {
		t.Fatal("returned the wrong items", finaloutput)
	}
	if finaloutput[0].Status.Conditions[0].LastTransitionTime != ocTime(tNow.Add(timestampHelper(16))) || finaloutput[1].Status.Conditions[0].LastTransitionTime != ocTime(tNow.Add(timestampHelper(25))) {
		t.Fatal("returned the wrong items", finaloutput)
	}

	retentionTime, _ = buildRetentionDuration(16)
	finaloutput = discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 2, t)

	retentionTime, _ = buildRetentionDuration(17)
	finaloutput = discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 1, t)

	retentionTime, _ = buildRetentionDuration(25)
	finaloutput = discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 1, t)

	retentionTime, _ = buildRetentionDuration(26)
	finaloutput = discardFailureTimestamps(testData, retentionTime)
	testResultLength(finaloutput, 0, t)

}
func testResultLength(res []jobItem, expected int, t *testing.T) {
	if len(res) != expected {
		t.Errorf("output is expected to be %d items %+v", expected, res)
	}
}

func timestampHelper(userVal int) time.Duration {
	retentionActual := "-" + strconv.Itoa(userVal) + "m"
	rDur, _ := time.ParseDuration(retentionActual)
	return rDur
}
