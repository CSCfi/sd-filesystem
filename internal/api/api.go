package api

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	// RequestTimeout dictates the number of seconds to wait before timing out an http request
	RequestTimeout int
	certPath       string
	metadataURL    string
	dataURL        string
	httpRetry      int
	token          string
	client         *http.Client
)

// Container stores container data from metadata API
type Container struct {
	Bytes int64  `json:"bytes"`
	Count int64  `json:"count"`
	Name  string `json:"name"`
}

// DataObject stores object data from metadata API
type DataObject struct {
	Bytes       int64  `json:"bytes"`
	ContentType string `json:"content_type"`
	Hash        string `json:"hash"`
	Name        string `json:"name"`
	Subdir      string `json:"subdir"`
}

func init() {
	certPath = os.Getenv("SD_CONNECT_CERTS")
	metadataURL = os.Getenv("SD_CONNECT_METADATA_API")
	dataURL = os.Getenv("SD_CONNECT_DATA_API")
	httpRetry = 5

	// init() has not been called in main yet so logger will be at level info.
	// Take this into consideration if more logs are added
	verifyURL(metadataURL, "SD_CONNECT_METADATA_API")
	verifyURL(dataURL, "SD_CONNECT_DATA_API")
}

func verifyURL(apiURL, env string) {
	if apiURL == "" {
		log.Fatalf("Environment variable %s not set", env)
	}
	// Verify that repository URL is valid
	u, err := url.ParseRequestURI(apiURL)
	if err != nil {
		log.Error(err)
		log.Fatalf("%s is not valid", env)
	}
	if u.Scheme != "https" {
		log.Fatalf("URL %s does not have scheme 'https'", apiURL)
	}
}

// CreateToken creates the authorization token based on username + password
func CreateToken(username, password string) {
	token = base64.StdEncoding.EncodeToString([]byte(username + ":" + string(password)))
}

// InitializeClient ...
func InitializeClient() error {
	// Handle certificate if one is set
	caCertPool := x509.NewCertPool()
	if len(certPath) > 0 {
		caCert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return err
		}
		caCertPool.AppendCertsFromPEM(caCert)
	} else {
		caCertPool = nil // So that default root certs are used
	}

	// Set up HTTP client
	timeout := time.Duration(RequestTimeout) * time.Second
	tr := http.DefaultTransport.(*http.Transport).Clone()

	tr.MaxConnsPerHost = 100
	tr.MaxIdleConnsPerHost = 100
	//tr.MaxIdleConns

	tr.TLSClientConfig = &tls.Config{
		RootCAs: caCertPool,
	}
	client = &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	_, err := client.Head(metadataURL)
	if err != nil {
		return err
	}

	// Temporarily disabling the keep-alive feature so that this unnecessary
	// connection does not take one of our goroutines
	tr.DisableKeepAlives = true
	_, err = client.Head(dataURL)
	if err != nil {
		return err
	}
	tr.DisableKeepAlives = false

	log.Debug("Initializing http client successful")
	return nil
}

// makeRequest builds an authenticated HTTP client
// which sends HTTP requests and parses the responses
func makeRequest(url string, query map[string]string, headers map[string]string) ([]byte, error) {
	var response *http.Response

	// Build HTTP request
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Place query params if they are set
	q := request.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()

	// Place mandatory headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Basic "+token)

	// Place additional headers if any are available
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	// Execute HTTP request
	// retry the request as specified by httpRetry variable
	for count := 0; count == 0 || (err != nil && count < httpRetry); {
		response, err = client.Do(request)
		log.Debugf("Trying Request %s, attempt %d/%d", request.URL, count+1, httpRetry)
		count++
	}
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 && response.StatusCode != 206 {
		return nil, fmt.Errorf("API responded with status %d", response.StatusCode)
	}

	// Parse request
	r, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Request %s returned a response", request.URL)
	return r, nil
}

// GetProjects gets all projects user has access to
func GetProjects() ([]string, error) {
	// Request projects
	response, err := makeRequest(strings.TrimRight(metadataURL, "/")+"/projects", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Retrieving projects failed:\n%w", err)
	}

	// Parse the JSON response into a slice
	var projects []string
	if err := json.Unmarshal(response, &projects); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response for retrieving projects: %w", err)
	}

	log.Info("Retrieved projects as per request")
	return projects, nil
}

// GetContainers gets conatiners inside the object
func GetContainers(project string) ([]Container, error) {
	// Request conteiners
	response, err := makeRequest(strings.TrimRight(metadataURL, "/")+"/project/"+
		url.QueryEscape(project)+"/containers", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Retrieving container from project %s failed: %w", project, err)
	}

	// Parse the JSON response into a slice
	var containers []Container
	if err := json.Unmarshal(response, &containers); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response for retrieving containers:\n%w", err)
	}

	log.Infof("Retrieved containers for project %s", project)
	return containers, nil
}

// GetObjects gets objects inside container
func GetObjects(project, container string) ([]DataObject, error) {
	// Request objects
	response, err := makeRequest(strings.TrimRight(metadataURL, "/")+"/project/"+
		url.QueryEscape(project)+"/container/"+url.QueryEscape(container)+"/objects", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("Retrieving objects from container %s failed: %w", container, err)
	}

	// Parse the JSON response into a struct
	var objects []DataObject
	if err := json.Unmarshal(response, &objects); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response for retrieving objects:\n%w", err)
	}
	log.Infof("Retrieved objects for %s", project+"/"+container)
	return objects, nil
}

// DownloadData gets content of object from data API
func DownloadData(path string, start int64, end int64) ([]byte, error) {
	parts := strings.SplitN(strings.TrimLeft(path, "/"), "/", 3)
	// Query params
	query := map[string]string{
		"project":   parts[0],
		"container": parts[1],
		"object":    parts[2],
	}

	// Additional headers
	headers := map[string]string{"Range": "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)}

	// Request data
	response, err := makeRequest(strings.TrimRight(dataURL, "/")+"/data", query, headers)
	if err != nil {
		return nil, err
	}
	log.Infof("Downloaded object %s from coordinates %d-%d", path, start, end-1)
	return response, nil
}
