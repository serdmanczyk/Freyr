package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/serdmanczyk/freyr/middleware"
	"github.com/serdmanczyk/freyr/models"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	//"net/http/httputil"
)

type ApiClient struct {
	client    *http.Client
	Domain    string
	UserEmail string
	Secret    models.Secret
}

func New(domain, userEmail, base64Secret string) (*ApiClient, error) {
	secret, err := models.SecretFromBase64(base64Secret)
	if err != nil {
		return nil, err
	}

	// TODO: remove this when we get a legit cert, or make optional?
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return &ApiClient{
		client:    client,
		Domain:    domain,
		UserEmail: userEmail,
		Secret:    secret,
	}, nil
}

func (c *ApiClient) GetLatest() ([]models.Reading, error) {
	var readings []models.Reading

	req, err := http.NewRequest("GET", c.Domain+"/api/latest", nil)
	if err != nil {
		return readings, err
	}

	middleware.SignRequest(c.Secret, c.UserEmail, req)
	resp, err := c.client.Do(req)
	if err != nil {
		return readings, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readings, errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

func (c *ApiClient) GetReadings(coreid string, start, end time.Time) ([]models.Reading, error) {
	var readings []models.Reading

	query := url.Values{}
	query.Add("start", start.Format(time.RFC3339))
	query.Add("end", end.Format(time.RFC3339))
	query.Add("core", coreid)
	reqUrl := c.Domain + "/api/readings?" + query.Encode()

	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return readings, err
	}

	middleware.SignRequest(c.Secret, c.UserEmail, req)
	resp, err := c.client.Do(req)
	if err != nil {
		return readings, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return readings, errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(readings)
	if err != nil {
		return readings, err
	}

	return readings, nil
}

func (c *ApiClient) PostReading(reading models.Reading) error {
	form := url.Values{}
	form.Set("event", "post_reading")
	form.Set("coreid", reading.CoreId)
	form.Set("published_at", reading.Posted.Format(models.JsonTime))
	form.Set("data", reading.DataJSON())

	formStr := form.Encode()
	reqBody := strings.NewReader(formStr)
	req, err := http.NewRequest("POST", c.Domain+"/api/reading", reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = int64(len(formStr))
	middleware.SignRequest(c.Secret, c.UserEmail, req)

	//bytes, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	return err
	//}
	//fmt.Println(string(bytes))

	resp, err := c.client.Do(req)
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

func (c *ApiClient) RotateSecret() (string, error) {
	req, err := http.NewRequest("POST", "/reading", nil)
	if err != nil {
		return "", err
	}

	middleware.SignRequest(c.Secret, c.UserEmail, req)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("Http Error: %d", resp.StatusCode))
	}

	newSecret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(newSecret), nil
}
