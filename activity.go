package workfusion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

const (
	MethodGET  = "GET"
	MethodPOST = "POST"
	MethodPUT  = "PUT"
)

// Activity is an activity that is used to invoke a REST Operation
// settings : {url, username, password}
// input    : {uuid}
// outputs  : {uuis, data}
type Activity struct {
	settings   *Settings
	client     *http.Client
	authTokens AuthTokens
}

type AuthTokens struct {
	CsrfToken      string
	CsrfHeaderName string
	JSESSIONID     string
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

func New(ctx activity.InitContext) (activity.Activity, error) {
	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	act := &Activity{settings: s}

	var err error
	act.client, err = getHttpClient(10)
	if err != nil {
		return nil, err
	}

	authTokens, err = connectToWF(client, s.URL, s.Username, s.Password)
	if err != nil {
		// replace password in error!!!!!
		return nil, err
	}

	return act, nil
}

func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Invokes a REST Operation
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return false, err
	}

	baseUrl := a.settings.URL

	logger := ctx.Logger()
	if logger.DebugEnabled() {
		logger.Debugf("WorkFusion Copy and Run: %s", input.UUID)
	}

	if logger.TraceEnabled() {
		logger.Trace("Response body:", result)
	}

	output := &Output{Status: resp.StatusCode, Data: result, Headers: respHeaders, Cookies: cookies}
	err = ctx.SetOutputObject(output)
	if err != nil {
		return false, err
	}

	return true, nil
}

////////////////////////////////////////////////////////////////////////////////////////
// Utils

func getHttpClient(timeout int) (http.Client, error) {

	client := &http.Client{}

	httpTransportSettings := &http.Transport{}

	if timeout > 0 {
		httpTransportSettings.ResponseHeaderTimeout = time.Second * time.Duration(timeout)
	}

	client.Transport = httpTransportSettings

	return *client, nil
}

func getRestResponse(client http.Client, method string, uri string, headers map[string]string, reqBody io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(method, uri, reqBody)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return resp, errors.New("Bad Response: " + resp.Status)
	}

	if resp == nil {
		return resp, errors.New("Empty Response")
	}

	return resp, nil
}

func getBodyAsText(respBody io.ReadCloser) string {

	defer func() {
		if respBody != nil {
			_ = respBody.Close()
		}
	}()

	var response = ""

	if respBody != nil {
		b := new(bytes.Buffer)
		b.ReadFrom(respBody)
		response = b.String()
	}

	return response
}

func getBodyAsJSON(respBody io.ReadCloser) (interface{}, error) {

	defer func() {
		if respBody != nil {
			_ = respBody.Close()
		}
	}()

	d := json.NewDecoder(respBody)
	d.UseNumber()
	var response interface{}
	err := d.Decode(&response)
	if err != nil {
		switch {
		case err == io.EOF:
			return nil, nil
		default:
			return nil, err
		}
	}

	return response, nil
}

func connectToWF(client http.Client, baseUrl string, username string, password string) (AuthTokens, error) {

	uri := baseUrl + fmt.Sprintf("/dologin?j_username=%s&j_password=%s", username, password)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	authTokens := AuthTokens{}

	resp, err := getRestResponse(client, MethodPOST, uri, headers, nil)
	if err != nil {
		return authTokens, err
	}

	response, err := getBodyAsJSON(resp.Body)
	if err != nil {
		return authTokens, err
	}

	loginResultMap := response.(map[string]interface{})
	success := loginResultMap["success"].(bool)
	if success {
		authTokens.CsrfHeaderName = loginResultMap["csrfHeaderName"].(string)
		authTokens.CsrfToken = loginResultMap["csrfToken"].(string)

		var cookieHeader string
		for _, header := range resp.Header["Set-Cookie"] {
			cookieHeader = header
		}

		cookies := strings.Split(cookieHeader, ";")
		if len(cookies) == 0 {
			return authTokens, errors.New("No cookies")
		}

		for _, cookie := range cookies {
			if strings.Index(cookie, "JSESSIONID=") != -1 {
				authTokens.JSESSIONID = cookie
				break
			}
		}

		if authTokens.JSESSIONID == "" {
			return authTokens, errors.New("JSESSIONID not found")
		}
	} else {
		return authTokens, errors.New("Login was not successful")
	}

	return authTokens, nil
}

func copyBP(client http.Client, baseUrl string, authTokens AuthTokens, UUID string) (string, error) {

	uri := baseUrl + "/v1/bp-instances/copy"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cookie"] = authTokens.JSESSIONID
	headers[authTokens.CsrfHeaderName] = authTokens.CsrfToken

	copyRequest := CopyBPRequest{
		DataCopy:                  true,
		DeepCopy:                  false,
		DeepCopySuffix:            "",
		IndependentDefinition:     false,
		IndependentDefinitionName: "",
		InstanceUUID:              UUID,
		ProcessCopy:               true,
	}
	reqBodyJSON, err := json.Marshal(copyRequest)
	if err != nil {
		return "", err
	}
	reqBody := bytes.NewBuffer([]byte(reqBodyJSON))

	resp, err := getRestResponse(client, MethodPOST, uri, headers, reqBody)
	if err != nil {
		return "", err
	}

	response, err := getBodyAsJSON(resp.Body)
	if err != nil {
		return "", err
	}

	responseMap := response.(map[string]interface{})

	if len(responseMap) == 0 {
		return "", errors.New("Empty responseMap")
	}

	return responseMap["result"].(string), nil
}

func runBP(client http.Client, baseUrl string, authTokens AuthTokens, UUID string) (string, error) {

	uri := baseUrl + fmt.Sprintf("/v1/bp-instances/%s/run", UUID)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cookie"] = authTokens.JSESSIONID
	headers[authTokens.CsrfHeaderName] = authTokens.CsrfToken

	resp, err := getRestResponse(client, MethodPUT, uri, headers, nil)
	if err != nil {
		return "", err
	}

	response := getBodyAsText(resp.Body)

	return response, nil
}

func checkRunStatus(client http.Client, baseUrl string, authTokens AuthTokens, UUID string) (bool, error) {

	uri := baseUrl + fmt.Sprintf("/v1/bp-instances/%s/reached-final-step", UUID)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cookie"] = authTokens.JSESSIONID
	headers[authTokens.CsrfHeaderName] = authTokens.CsrfToken

	resp, err := getRestResponse(client, MethodGET, uri, headers, nil)
	if err != nil {
		return false, err
	}

	response := getBodyAsText(resp.Body)

	return (response == "true"), nil
}

func fetchResults(client http.Client, baseUrl string, authTokens AuthTokens, UUID string) (string, error) {

	uri := baseUrl + fmt.Sprintf("/v1/bp-instances/%s/results", UUID)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Cookie"] = authTokens.JSESSIONID
	headers[authTokens.CsrfHeaderName] = authTokens.CsrfToken

	reqBody := bytes.NewBuffer([]byte("{}"))

	resp, err := getRestResponse(client, MethodPOST, uri, headers, reqBody)
	if err != nil {
		return "", err
	}

	response := getBodyAsText(resp.Body)

	return response, nil
}
