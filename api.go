package dogdirect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var endpointv1 = "https://api.datadoghq.com/api/v1"

type API struct {
	apikey  string
	appkey  string
	timeout time.Duration
}

func NewAPI(apikey string, appkey string, timeout time.Duration) API {
	return API{
		apikey:  apikey,
		appkey:  appkey,
		timeout: timeout,
	}
}

func (a API) AddPoints(metrics []*Metric) error {

	post := map[string][]*Metric{
		"series": metrics,
	}
	endpoint := fmt.Sprintf("%s/series?api_key=%s", endpointv1, a.apikey)

	return write(endpoint, post, a.timeout)
}

// AddHostTags adds a host tags to a given host.
// If host is new or doesn't have metrics yet, this call will fail.
func (a API) AddHostTags(host string, source string, tags []string) error {
	if source == "" {
		source = "user"
	}
	post := map[string][]string{
		"tags": tags,
	}
	endpoint := fmt.Sprintf("%s/tags/hosts/%s?api_key=%s&application_key=%s&source=%s", endpointv1, host, a.apikey, a.appkey, source)

	return write(endpoint, post, a.timeout)
}

// writes a json blob
func write(endpoint string, data interface{}, timeout time.Duration) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: timeout,
	}

	body := bytes.NewReader(raw)
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok {
			// if the error has the url in it,
			// then retrieve the inner error
			// and ditch the url (which might contain secrets)
			err = urlErr.Err
		}
		return err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted, http.StatusCreated:
		return nil
	}

	return fmt.Errorf("http status %v: %s", resp.StatusCode, string(responseBody))
}
