// Package client is a convenience package that provides methods to interface
// with the Freyr server's HTTP API.
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/serdmanczyk/freyr/middleware"
	"github.com/serdmanczyk/freyr/models"
	"github.com/serdmanczyk/freyr/oauth"
	"io"
	"io/ioutil"
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
		return readings, fmt.Errorf("Http Error: %d", resp.StatusCode)
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
		return readings, fmt.Errorf("Http Error: %d", resp.StatusCode)
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
		return fmt.Errorf("Http Error: %d", resp.StatusCode)
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

	//bytes, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(string(bytes))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Http Error %d: %s", resp.StatusCode, string(bytes))
	}

	return nil
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
		return models.Secret([]byte{}), fmt.Errorf("Http Error: %d", resp.StatusCode)
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
		return nilSecret, fmt.Errorf("Http Error: %d", resp.StatusCode)
	}

	var dest bytes.Buffer

	decoder := base64.NewDecoder(base64.URLEncoding, resp.Body)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		return nilSecret, err
	}

	return models.Secret(dest.Bytes()), nil
}
