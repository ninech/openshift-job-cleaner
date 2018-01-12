package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var userConfig *config

func init() {
	userConfig = new(config)
}

func main() {

	doInitialSetup()

	ocAddr := buildOcLoginAddress(os.Getenv("KUBERNETES_PORT_443_TCP_ADDR"), os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS"))
	err := ocLogin(ocAddr)
	panicOnError(err)

	//get all the projects in the cluster that this sa has access to
	projectBuff, err := callOc("projects", "-q")
	panicOnError(err)
	projects := strings.Split(projectBuff.String(), "\n")
	projects = projects[:len(projects)-1]
	if len(projects) < 1 {
		infoAndExit("no accessible projects found in cluster, exiting", 0)
	}

	//iterate the projects returned
	for _, v := range projects {

		//if blacklisted skip
		if userConfig.namespaceBlacklisted(v) {
			log.Println("skipping blacklisted namespace:", v)
			continue
		}

		jobstate, err := getJobsInNamespace(v)
		if continueOnError(err) {
			log.Println("skipping due to error getting jobs in namespace:", v)
			continue
		}

		// Get the descriptor for this job
		descriptor := userConfig.getJobDescriptor(v)

		// sort jobs by status, return TRUE for job success
		successes, failures := sort(jobstate.Items, func(item jobItem) bool {
			return !(item.Status.Failed > 0)
		})
		log.Println(failures)

		// SUCCESSES
		// retention is checked by looking at the completion date. if it is newer than now - retentionPeriod remove it from the array of jobs to delete
		err = doClean(descriptor.Success.Retention, v, successes, discardSuccessTimestamps)
		if continueOnError(err) {
			continue
		}

		// FAILURES
		err = doClean(descriptor.Failure.Retention, v, failures, discardFailureTimestamps)
		if continueOnError(err) {
			continue
		}
	}
}

// doClean takes the job items, filters them with filter, executes the call to oc to delete them, and prints the output if any
func doClean(retention int, namespace string, jobItems []jobItem, filter func(items []jobItem, earliest time.Time) []jobItem) error {
	retentionTime, err := buildRetentionDuration(retention)
	if err != nil {
		return err
	}
	jobItems = filter(jobItems, retentionTime)
	if len(jobItems) > 0 {
		output, err := callOc(buildDeletionString(jobItems, namespace)...)
		if err != nil {
			return err
		}
		if output.Len() > 0 {
			log.Println(output.String())
		}
	}
	return nil
}

// doInitialSetup sets up logging and reads in the config
// this cannot be in an init block as then it will need to work during testing as well
func doInitialSetup() {
	err := initLogging(os.Getenv("SENTRY_DSN"))
	panicOnError(err)
	confPath := os.Getenv("OJC_CONFIG_PATH")
	if len(confPath) == 0 {
		confPath = "/opt/ojc/ojc.yml"
	}
	userConfig, err = loadConfig(confPath) //userConfig is a global
	panicOnError(err)
}

// buildDeletionString assembles a slice of strings that will be passed to callOc to execute a job deletion command
func buildDeletionString(jobs []jobItem, namespace string) []string {
	jobStrings := make([]string, 1)
	jobStrings[0] = "delete"
	jobStrings = append(jobStrings, "-n")
	jobStrings = append(jobStrings, namespace)
	for _, v := range jobs {
		jobStrings = append(jobStrings, "job/"+v.Metadata.Name)
	}
	return jobStrings
}

// getFileAsString opens a file at path and returns its contents as a string
// used for extracting kube secrets from the runtime container
func getFileAsString(filePath string) (string, error) {
	tokBytes, err := ioutil.ReadFile(filePath)
	return string(tokBytes), err
}
