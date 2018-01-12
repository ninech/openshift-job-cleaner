package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func init() {
	ocDebug = true
}

func TestBuildOcLoginAddress(t *testing.T) {
	loginAddr := buildOcLoginAddress("trashbat.co.ck", "443")
	testPrefix(loginAddr, "https://", t)
	testSuffix(loginAddr, ":443", t)

	loginAddr = buildOcLoginAddress("http://trashbat.co.ck", "1234")
	testPrefix(loginAddr, "http://", t)
	testSuffix(loginAddr, ":1234", t)

	loginAddr = buildOcLoginAddress("https://trashbat.co.ck", "666")
	testPrefix(loginAddr, "https://", t)
	testSuffix(loginAddr, ":666", t)

}

func TestGetJobsInNamespace(t *testing.T) {
	execCommand = fakeGetJobsCommand
	jobs, err := getJobsInNamespace("my-namespace")
	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if len(jobs.Items) != 1 {
		t.Errorf("expected 1 job got %d", len(jobs.Items))
	}
}

func TestGetJobsInNamespaceError(t *testing.T) {
	execCommand = fakeFailingGetJobsCommand
	_, err := getJobsInNamespace("garbageTown")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestOcLoginTokenError(t *testing.T) {
	os.Setenv("KUBERNETES_SA_TOKEN_PATH", "fail")
	execCommand = fakeLogincommandNoToken
	err := ocLogin("garbageTown")
	if err == nil {
		t.Error("Expected error when token does not exist")
	}
}

func TestOcLogin(t *testing.T) {
	execCommand = fakeLoginCommand
	os.Setenv("KUBERNETES_SA_TOKEN_PATH", "./test/faketoken")
	err := ocLogin("garbageTown")
	if err != nil {
		t.Error(err)
	}
	os.Unsetenv("KUBERNETES_SA_TOKEN_PATH")
}

func TestCallOc(t *testing.T) {
	execCommand = fakePassingExecCommand
	out, err := callOc("")

	if err != nil {
		t.Errorf("Expected nil error, got %#v", err)
	}
	if out.String() != PassingOcCallOutput {
		t.Errorf("Expected %q, got %q", PassingOcCallOutput, out)
	}
}

func TestCallOcFail(t *testing.T) {

	execCommand = fakeFailingExecCommand
	_, err := callOc("")

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	if !strings.HasPrefix(err.Error(), FailingOcCallOutput) {
		t.Fatal("error should start with output written to command stdErr")
	}

}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("GO_TEST_FAIL") == "1" {
		fmt.Fprintf(os.Stderr, "%s", FailingOcCallOutput)
		os.Exit(0)
	} else if os.Getenv("GO_NAMESPACE_GET_TEST") == "1" {
		jobstate := new(currentJobstate)
		jobstate.Items = make([]jobItem, 1)
		jobstate.Items[0] = jobItem{}
		data, err := yaml.Marshal(jobstate)
		if err != nil {
			t.Fatal("marshalling something set up in a test should not fail, what did you do!")
		}
		fmt.Fprintf(os.Stdout, "%s", data)
		os.Exit(0)
	} else if os.Getenv("GO_NAMESPACE_GET_TEST_FAIL") == "1" {
		os.Exit(1)
	} else {
		fmt.Fprintf(os.Stdout, "%s", PassingOcCallOutput)
		os.Exit(0)
	}

}

//Helpers
const PassingOcCallOutput = "test passed"
const FailingOcCallOutput = "test failed"

func testPrefix(text string, prefix string, t *testing.T) {
	if !strings.HasPrefix(text, prefix) {
		t.Fatal("prefix did not match expected", "expected:", prefix, "string:", text)
	}
}

func testSuffix(text string, suffix string, t *testing.T) {
	if !strings.HasSuffix(text, suffix) {
		t.Fatal("suffix did not match expected", "expected:", suffix, "string:", text)
	}
}

func fakePassingExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeFailingExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_TEST_FAIL=1"}
	return cmd
}

func fakeGetJobsCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_NAMESPACE_GET_TEST=1"}
	return cmd
}

func fakeFailingGetJobsCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1", "GO_NAMESPACE_GET_TEST_FAIL=1"}
	return cmd
}

func fakeLoginCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func fakeLogincommandNoToken(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}
