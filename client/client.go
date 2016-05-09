package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	nilSecret              = models.Secret([]byte{})
	client    *http.Client = nil
)

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

type Signator interface {
	Sign(r *http.Request)
}

type WebSignator struct {
	Token string
}

func (s WebSignator) Sign(r *http.Request) {
	r.Header.Add("Cookie", oauth.CookieName+"="+s.Token)
}

type ApiSignator struct {
	UserEmail string
	Secret    models.Secret
}

func NewApiSignator(userEmail, base64Secret string) (*ApiSignator, error) {
	secret, err := models.SecretFromBase64(base64Secret)
	if err != nil {
		return nil, err
	}

	return &ApiSignator{
		UserEmail: userEmail,
		Secret:    secret,
	}, nil
}

func (s ApiSignator) Sign(r *http.Request) {
	middleware.SignRequest(s.Secret, s.UserEmail, r)
}

type DeviceSignator struct {
	UserEmail string
	Token     string
}

func (s *DeviceSignator) Sign(r *http.Request) {
	r.Header.Add(middleware.AuthTypeHeader, middleware.DeviceAuthTypeValue)
	r.Header.Add(middleware.AuthUserHeader, s.UserEmail)
}

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
		return readings, errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

func GetReadings(s Signator, domain, coreid string, start, end time.Time) ([]models.Reading, error) {
	var readings []models.Reading

	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", end.Format(time.RFC3339))
	query.Add("core", coreid)
	reqUrl := domain + "/api/readings?" + query.Encode()

	req, err := http.NewRequest("GET", reqUrl, nil)
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
		return readings, errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

func DeleteReadings(s Signator, domain, coreid string, start, end time.Time) error {
	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", end.Format(time.RFC3339))
	query.Add("core", coreid)
	reqUrl := domain + "/api/delete_readings?" + query.Encode()

	req, err := http.NewRequest("DELETE", reqUrl, nil)
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
		return errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	return nil
}

func PostReading(s Signator, domain string, reading models.Reading) error {
	form := url.Values{}
	form.Set("event", "post_reading")
	form.Set("coreid", reading.CoreId)
	form.Set("published_at", reading.Posted.Format(models.JsonTime))
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
		return errors.New(fmt.Sprintf("Http Error %d: %s", resp.StatusCode, string(bytes)))
	}

	return nil
}

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
		return models.Secret([]byte{}), errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	var dest bytes.Buffer

	decoder := base64.NewDecoder(base64.URLEncoding, resp.Body)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		return nilSecret, err
	}

	return models.Secret(dest.Bytes()), nil
}

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
		return nilSecret, errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	var dest bytes.Buffer

	decoder := base64.NewDecoder(base64.URLEncoding, resp.Body)
	_, err = io.Copy(&dest, decoder)
	if err != nil {
		return nilSecret, err
	}

	return models.Secret(dest.Bytes()), nil
}
