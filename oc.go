package main

import (
	"bytes"
	"errors"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"os/exec"
	"strings"
)

var ocDebug = false

// ocLogin builds the login command arguments and executes it
func ocLogin(ocAddr string) error {
	saTokenPath := os.Getenv("KUBERNETES_SA_TOKEN_PATH")
	if saTokenPath == "" {
		saTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint
	}
	certAuthorityPath := os.Getenv("KUBERNETES_CERT_AUTHORITY_PATH")
	if certAuthorityPath == "" {
		certAuthorityPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt" //nolint
	}
	token, err := getFileAsString(saTokenPath)
	if err != nil {
		return err
	}
	_, err = callOc("login", ocAddr, "--token", token, "--certificate-authority", certAuthorityPath)
	return err
}

// buildOcLoginAddress is a simple helper to construct a valid kube endpoint address and ensure it has a protocol at the start
func buildOcLoginAddress(kubeAddr string, kubePort string) string {
	varString := kubeAddr + ":" + kubePort
	if !strings.HasPrefix(varString, "http") {
		//test for any http prefix and if none default to https
		varString = "https://" + varString
	}
	return varString
}

// getJobsInNamespace retrieves the currentJobstate for a given namespace
func getJobsInNamespace(namespace string) (*currentJobstate, error) {
	//get all the jobs in the namespace
	jobsBuff, err := callOc("get", "jobs", "-n", namespace, "-o", "yaml")
	if err != nil {
		return new(currentJobstate), err
	}
	jobstate := new(currentJobstate)
	err = yaml.Unmarshal(jobsBuff.Bytes(), jobstate)
	return jobstate, err
}

// execCommand var helps us to mock out commands in testing
var execCommand = exec.Command

// callOc is the proxy for all exec calls to oc and deals with potential output errors, bubbling them up.
func callOc(args ...string) (bytes.Buffer, error) {
	if ocDebug {
		log.Println("oc", args)
	}
	cmd := execCommand("oc", args...)
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if errBuf.Len() > 0 || err != nil {
		//we count anything that writes to stderr as a failure
		errString := errBuf.String()
		if err != nil {
			errString += err.Error()
		}
		return outBuf, errors.New(errString)
	}
	return outBuf, err
}
