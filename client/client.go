// Package client is a convenience package that provides methods to interface
// with the Freyr server's HTTP API.
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/serdmanczyk/bifrost"
	"github.com/serdmanczyk/freyr/middleware"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/oauth"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	nilSecret = models.Secret([]byte{})
	client    *http.Client
)

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// Signator is an interface representing a type that is capable of signing
// requests.
type Signator interface {
	Sign(r *http.Request)
}

// WebSignator is used to sign requests in the same way they would
// be when a user is accessing the server through the web interface.
type WebSignator struct {
	Token string
}

// Sign is used to sign an http.Request by adding the signed token
// (signed by the server's key) as a header.
func (s WebSignator) Sign(r *http.Request) {
	r.Header.Add("Cookie", oauth.CookieName+"="+s.Token)
}

// APISignator is used to sign requests in the method prescribed for API calls.
type APISignator struct {
	UserEmail string
	Secret    models.Secret
}

// NewAPISignator generates a new ApiSignator, conveniencing decoding the
// secret from base64.
func NewAPISignator(userEmail, base64Secret string) (*APISignator, error) {
	secret, err := models.SecretFromBase64(base64Secret)
	if err != nil {
		return nil, err
	}

	return &APISignator{
		UserEmail: userEmail,
		Secret:    secret,
	}, nil
}

// Sign signs an http.Request by applying an API signature.
func (s APISignator) Sign(r *http.Request) {
	middleware.SignRequest(s.Secret, s.UserEmail, r)
}

// DeviceSignator is used to sign a request in the way prescribed for device
// requests.
type DeviceSignator struct {
	UserEmail string
	Token     string
}

// Sign (incomplete) signs a request by applying the headers indicating which
// device signed the request for which user and providing a token signed with
// that user's secret.
func (s *DeviceSignator) Sign(r *http.Request) {
	r.Header.Add(middleware.AuthTypeHeader, middleware.DeviceAuthTypeValue)
	r.Header.Add(middleware.AuthUserHeader, s.UserEmail)
}

func responseError(resp *http.Response) error {
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	return fmt.Errorf("Http Error %d: %s", resp.StatusCode, string(bodyBytes))
}

// GetLatest gets the latest readings, per device, for a user.
func GetLatest(s Signator, domain string) ([]models.Reading, error) {
	var readings []models.Reading

	req, err := http.NewRequest("GET", domain+"/api/latest", nil)
	if err != nil {
		return readings, err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return readings, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readings, responseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

// GetReadings gets all the readings stored in a specified time frame for a user.
func GetReadings(s Signator, domain, coreid string, start, end time.Time) ([]models.Reading, error) {
	var readings []models.Reading

	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", end.Format(time.RFC3339))
	query.Add("core", coreid)
	reqURL := domain + "/api/readings?" + query.Encode()

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return readings, err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return readings, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readings, responseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

// DeleteReadings deletes all readings within a specified time frame.
func DeleteReadings(s Signator, domain, coreid string, start, end time.Time) error {
	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", end.Format(time.RFC3339))
	query.Add("core", coreid)
	reqURL := domain + "/api/delete_readings?" + query.Encode()

	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return responseError(resp)
	}

	return nil
}

// PostReading posts a new reading
func PostReading(s Signator, domain string, reading models.Reading) error {
	form := url.Values{}
	form.Set("event", "post_reading")
	form.Set("coreid", reading.CoreID)
	form.Set("published_at", reading.Posted.Format(models.JSONTime))
	form.Set("data", reading.DataJSON())

	formStr := form.Encode()
	reqBody := strings.NewReader(formStr)
	req, err := http.NewRequest("POST", domain+"/api/reading", reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = int64(len(formStr))
	s.Sign(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return responseError(resp)
	}

	return nil
}

// PostReadings posts a list of readings
func PostReadings(s Signator, domain string, readings []models.Reading) (string, error) {
	reqBody := new(bytes.Buffer)
	err := json.NewEncoder(reqBody).Encode(&readings)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", domain+"/api/readings", reqBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	s.Sign(req)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", responseError(resp)
	}

	jobIdBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(jobIdBytes), nil
}

// GetSecret requests the system to generate a new secret for the user.
func GetSecret(s Signator, domain string) (models.Secret, error) {
	req, err := http.NewRequest("GET", domain+"/api/secret", nil)
	if err != nil {
		return nilSecret, err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return nilSecret, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.Secret([]byte{}), responseError(resp)
	}

	var dest bytes.Buffer

	decoder := base64.NewDecoder(base64.URLEncoding, resp.Body)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		return nilSecret, err
	}

	return models.Secret(dest.Bytes()), nil
}

// RotateSecret requests the system rotate a user's secret.  Request must be API signed.
func RotateSecret(s Signator, domain string) (models.Secret, error) {
	req, err := http.NewRequest("POST", domain+"/api/rotate_secret", nil)
	if err != nil {
		return nilSecret, err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return nilSecret, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nilSecret, responseError(resp)
	}

	var dest bytes.Buffer

	decoder := base64.NewDecoder(base64.URLEncoding, resp.Body)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		return nilSecret, err
	}

	return models.Secret(dest.Bytes()), nil
}

func GetJobStatus(s Signator, domain, jobID string) (*bifrost.JobStatus, error) {
	req, err := http.NewRequest("GET", domain+"/api/job?jobID="+jobID, nil)
	if err != nil {
		return nil, err
	}

	s.Sign(req)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, responseError(resp)
	}

	jobStatus := new(bifrost.JobStatus)
	err = json.NewDecoder(resp.Body).Decode(jobStatus)
	if err != nil {
		return nil, err
	}

	return jobStatus, nil
}

func WaitForJob(s Signator, domain, jobID string, timeout time.Duration) error {
	done := make(chan bool)
	failure := make(chan error)
	start := time.Now()
	deadline := start.Add(timeout)
	go func() {
		for {
			status, err := GetJobStatus(s, domain, jobID)
			if err != nil {
				failure <- err
				return
			}
			if status.Complete {
				done <- true
				return
			}
			if time.Now().After(deadline) {
				return
			}
			<-time.After(time.Millisecond * 100)
		}
	}()
	select {
	case <-done:
		log.Printf("Waited %s on Job %s", time.Now().Sub(start), jobID)
		return nil
	case err := <-failure:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timed out waiting for job to complete")
	}
}
