package main

import (
	"io/ioutil"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetFileAsString(t *testing.T) {

	rand.Seed(time.Now().Unix())
	for i := 0; i < 100; i++ {
		writeData, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal("cannot create temp file for test run")
		}
		defer writeData.Close()

		length := rand.Intn(letterIdxMax)
		stringData := RandStringBytesMaskImprSrc(length)
		writeData.WriteString(stringData)

		fileString, err := getFileAsString(writeData.Name())
		if err != nil {
			t.Fatal(err)
		}
		if fileString != stringData {
			t.Fatal("string data does not match expected output", "Expected:", stringData, "actual:", fileString)
		}
	}
}

func TestGetDescriptorForJob(t *testing.T) {

	defaultDescriptor := namespaceCleanupDescriptor{cleanupDescriptor{1}, cleanupDescriptor{2}}
	userConfig.Default = defaultDescriptor

	specialized := make(map[string]namespaceCleanupDescriptor, 2)
	mns := namespaceCleanupDescriptor{cleanupDescriptor{21}, cleanupDescriptor{22}}
	specialized["my-namespace"] = mns
	specialized["some-project"] = namespaceCleanupDescriptor{cleanupDescriptor{1}, cleanupDescriptor{24}}
	userConfig.Namespaces = specialized

	mynsD := userConfig.getJobDescriptor("my-namespace")
	if !reflect.DeepEqual(mynsD, &mns) {
		t.Fatal("namespace should return custom descriptor")
	}

	mynsD = userConfig.getJobDescriptor("testing-something")
	if !reflect.DeepEqual(mynsD, &defaultDescriptor) {
		t.Fatal("unknown namespace should return default descriptor")
	}

}

func TestBuildDeletionString(t *testing.T) {

	ja := jobItem{}
	ja.Metadata.Name = "test-1"
	jb := jobItem{}
	jb.Metadata.Name = "test-2"
	jc := jobItem{}
	jc.Metadata.Name = "test-3"
	jbs := []jobItem{ja, jb, jc}

	deleteStrings := buildDeletionString(jbs, "my-ns")
	expected := "delete -n my-ns job/test-1 job/test-2 job/test-3"
	deleteString := strings.Join(deleteStrings, " ")
	if deleteString != expected {
		t.Fatal("job string not built as expected", "Expected:", expected, "Actual:", deleteString)
	}
}

//HELPER FUNCTIONS

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
