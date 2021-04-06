package telemetry

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"time"
	"crypto/tls"
	"io/ioutil"
//  	"strings"
	"net/http"
	"net/url"
	"encoding/json"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

// eventType
const eventType = "library-os-ng"

// actionName
const actionName = "Piper Library OS"

// LibraryRepository that is passed into with -ldflags
var LibraryRepository string

// SiteID ...
var SiteID string

var disabled bool
var client piperhttp.Sender

// Initialize sets up the base telemetry data and is called in generated part of the steps
func Initialize(telemetryDisabled bool, stepName string) {
	disabled = telemetryDisabled

	// skip if telemetry is dieabled
	if disabled {
		log.Entry().Info("Telemetry reporting deactivated")
		return
	}

	if client == nil {
		client = &piperhttp.Client{}
	}

	client.SetOptions(piperhttp.ClientOptions{MaxRequestDuration: 5 * time.Second})

	if len(LibraryRepository) == 0 {
		LibraryRepository = "https://github.com/n/a"
	}

	if len(SiteID) == 0 {
		SiteID = "827e8025-1e21-ae84-c3a3-3f62b70b0130"
	}

	baseData = BaseData{
		URL:             LibraryRepository,
		ActionName:      actionName,
		EventType:       eventType,
		StepName:        stepName,
		SiteID:          SiteID,
		PipelineURLHash: getPipelineURLHash(), // http://server:port/jenkins/job/foo/
		BuildURLHash:    getBuildURLHash(),    // http://server:port/jenkins/job/foo/15/
	}
	//ToDo: register Logrus Hook

	// httpHook := &log.HTTPHook{CorrelationID: GeneralConfig.CorrelationID, pipelineURLHash: getPipelineURLHash(), buildURLHash: getBuildURLHash()}
	// log.RegisterHook(httpHook)

}

func getPipelineURLHash() string {
	return toSha1OrNA(os.Getenv("JOB_URL"))
}

func getBuildURLHash() string {
	return toSha1OrNA(os.Getenv("BUILD_URL"))
}

func toSha1OrNA(input string) string {
	if len(input) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%x", sha1.Sum([]byte(input)))
}

// SWA baseURL
const baseURL = "https://webanalytics.cfapps.eu10.hana.ondemand.com"

// SWA endpoint
const endpoint = "/tracker/log"

// Send ...
func Send(customData *CustomData) {
	data := Data{
		BaseData:     baseData,
		BaseMetaData: baseMetaData,
		CustomData:   *customData,
	}
	fmt.Println("Inside telemetry.go")
	// skip if telemetry is dieabled
	if disabled {
		return
	}

	request, _ := url.Parse(baseURL)
	request.Path = endpoint
	request.RawQuery = data.toPayloadString()
	log.Entry().WithField("request", request.String()).Debug("Sending telemetry data")

	SendDataToSplunk(customData)

	client.SendRequest(http.MethodGet, request.String(), nil, nil, nil)
	
	//request, _ := url.Parse("https://cm-poc-api-x5t2y6sjxq-uc.a.run.app")
	//request.Path = "/piper"
	//request.RawQuery = data.toPayloadString()
	//log.Entry().WithField("request", request.String()).Debug("Sending data to GCP")
	//client.SendRequest(http.MethodPost, request.String(), nil, nil, nil)	
}

type Event struct {
	//Time       interface{} `json:"time"`               // when the event happened
	Host       string  `json:"host"`                 // hostname
	Source     string  `json:"source,omitempty"`     // optional description of the source of the event; typically the app's name
	SourceType string  `json:"sourcetype,omitempty"` // optional name of a Splunk parsing configuration; this is usually inferred by Splunk
	Index      string  `json:"index,omitempty"`      // optional name of the Splunk index to store the event in; not required if the token has a default index set in Splunk
	Payload    MonitoringData `json:"event,omitempty"`      // throw any useful key/val pairs here
}

func SendDataToSplunk( customData *CustomData) {

	fmt.Println("Inside Splunk HTTP Method")

	data := MonitoringData{
		PipelineUrlHash: getPipelineURLHash(),
		BuildUrlHash: getBuildURLHash(),
		// StageName: customData.e_10,
		StepName: baseData.StepName,
		ExitCode: customData.ErrorCode,
		Duration: customData.Duration,
		ErrorCode: customData.ErrorCode,
		ErrorCategory: customData.ErrorCategory,
	}

	// data := Data{
	// 	BaseData:     baseData,
	// 	BaseMetaData: baseMetaData,
	// 	CustomData:   *customData,
	// }

	details := Event{
		Host:       getPipelineURLHash(),
		Index:      "cicd_pipeline_mon",
		SourceType: "_json",
		Payload:    data,
	}

	payload, err := json.Marshal(details)

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	splunkClient := &http.Client{}
	// fmt.Println(data.toPayloadString())
	
	fmt.Println("Data")
	mar, _ := json.Marshal(data)
	fmt.Println(string(mar))
	
	// mar, err = json.Marshal(baseMetaData)
	
	// fmt.Println("Base Meta Data")
	// fmt.Println(string(mar))
	
	// fmt.Println("Custom Data")
	// mar, err = json.Marshal(customData)
	// fmt.Println(string(mar))
	
	req, err := http.NewRequest("POST", "https://devsplunk-test-hc.splunk.devint.net.sap/services/collector/event", bytes.NewBuffer(payload))

	if err != nil {
		fmt.Println(err)
		// return errors.New("Error Sending Telemetry data to Splunk")
	}

	req.Header.Add("Authorization", "Splunk 445bcb88-4458-447c-a9fd-21c1026f4a75")
	req.Header.Add("Content-Type", "application/json")

	res, err := splunkClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
}
