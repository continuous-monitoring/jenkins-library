package log

import (
	"bytes"
	"encoding/json"
	"net/http"
	"fmt"
	// piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// HTTPHook provides a logrus hook which sends log data to a HTTP endpoint
// This can be used to provide data about the pipeline run for better monitoring
type HTTPHook struct {
	CorrelationID   string
	PipelineURLHash string
	BuildURLHash    string
}

// Levels returns the supported log level of the hook.
func (f *HTTPHook) Levels() []logrus.Level {
	fmt.Println("Adding required levels")
	return []logrus.Level{logrus.DebugLevel}
}

// Fire creates a new event from the error and sends it to http endpoint
func (f *HTTPHook) Fire(entry *logrus.Entry) error {
 	fmt.Println("Executing the http hook method to send data")
	details := entry.Data
	if details == nil {
		details = logrus.Fields{}
	}

	details["pipelineURLHash"] = f.PipelineURLHash
	details["buildURLHash"] = f.BuildURLHash

	reqBody, _ := json.Marshal(details)

	httpClient := &http.Client{}

	// header := make(http.Header)
	// header.Set("Content-Type", "application/json")

	url := "https://cm-poc-api-x5t2y6sjxq-uc.a.run.app/piper"
	// resp, httpErr := httpClient.SendRequest("POST", url, bytes.NewBuffer(details), header, nil)

	//try http as piperhttp has a cyclic dependency
	request, httpErr := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))

	request.Header.Set("Content-Type", "application/json")

	resp, httpErr := httpClient.Do(request)

	// utils.Stdout(reqBody)

	if httpErr != nil {
		return errors.Wrap(httpErr, "could not log data")
	} else if resp == nil {
		return errors.New("logging failed: did not retrieve a HTTP response")
	}

	return nil

}
