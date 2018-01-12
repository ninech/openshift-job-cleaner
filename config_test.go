package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestUnMarshall(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	file, err := ioutil.ReadFile(wd + string(os.PathSeparator) + "test" + string(os.PathSeparator) + "jobdata.yml")
	if err != nil {
		t.Fatal(err)
	}
	jobstate := new(currentJobstate)
	err = yaml.Unmarshal(file, jobstate)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnMarshallInvalid(t *testing.T) {
	jobstate := new(currentJobstate)
	file := `apiVersion: v1
items:
- apiVersion: batch/v1
  kind: Job
  status:
    completionTime: 2017-12-20T12:49:15Z
    conditions:
    - lastTransitionTime: 2017-12-20T12:49:15Z
    succeeded: 1`
	err := yaml.Unmarshal([]byte(file), jobstate)
	if err != nil {
		t.Fatal(err)
	}
	file = `apiVersion: v1
items:
- apiVersion: batch/v1
  kind: Job
  status:
    completionTime: totalinvalidgarbage
    conditions:
    - lastTransitionTime: also-invalid
    succeeded: 1`
	err = yaml.Unmarshal([]byte(file), jobstate)
	if err == nil {
		t.Fatal("invalid dates should raise an error")
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	_, err := loadConfig("blah/blah/shtonshtoenssone/ocj.yml")
	if err == nil {
		t.Fatal("invalid paths should return an error when attempling to load")
	}
}

func TestLoadConfig(t *testing.T) {

	expectedCfg := new(config)
	expectedCfg.Default.Success = cleanupDescriptor{MaxAge: 24, Retention: 1}
	expectedCfg.Default.Failure = cleanupDescriptor{MaxAge: 12, Retention: 5}
	expectedCfg.Blacklist = append(expectedCfg.Blacklist, "openshift-infra")

	nscld := namespaceCleanupDescriptor{cleanupDescriptor{MaxAge: 23, Retention: 6}, cleanupDescriptor{MaxAge: 18, Retention: 5}}
	expectedCfg.Namespaces = make(map[string]namespaceCleanupDescriptor)
	expectedCfg.Namespaces["my-namespace"] = nscld

	marshalled, err := yaml.Marshal(expectedCfg)
	if err != nil {
		t.Fatal("could not marshall test data into file")
	}
	writeData, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("cannot create temp file for test run")
	}
	defer writeData.Close()

	name := writeData.Name()
	n, err := writeData.Write(marshalled)
	if n == 0 || err != nil {
		t.Fatal("could not write test data")
	}

	ucfg, err := loadConfig(name)
	if err != nil {
		t.Fatal(err)
	}
	if len(ucfg.Blacklist) != 1 {
		t.Error("did not return correct config details from file")
	}

	if !reflect.DeepEqual(ucfg, expectedCfg) {
		t.Fatal("loaded config does not match saved config", ucfg, expectedCfg)
	}
}

func TestNamespaceBlackListed(t *testing.T) {
	expectedCfg := new(config)
	expectedCfg.Blacklist = append(expectedCfg.Blacklist, "openshift-infra")
	expectedCfg.Blacklist = append(expectedCfg.Blacklist, "whatever")

	if expectedCfg.namespaceBlacklisted("my-project") {
		t.Fatal("non blacklisted namespace should not return as blacklisted")
	}
	if !expectedCfg.namespaceBlacklisted("whatever") {
		t.Fatal("blacklisted namespace should not return as un-blacklisted")
	}
	if !expectedCfg.namespaceBlacklisted("openshift-infra") {
		t.Fatal("blacklisted namespace should not return as un-blacklisted")
	}

}
