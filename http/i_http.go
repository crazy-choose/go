package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/crazy-choose/go/log"
)

var impl *IHttpClient

type IHttpClient struct {
	apiHost string
	client  http.Client
}

func NewHttpClient(host string, timeout time.Duration) *IHttpClient {
	if host == "" {
		return nil
	}
	impl = &IHttpClient{
		apiHost: host,
		client: http.Client{
			Timeout: timeout,
		},
	}
	return impl
}

func (impl *IHttpClient) SendGetJson(ctx context.Context, url string, body []byte, internalClientRes interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, impl.apiHost+url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, internalClientRes)
	return err
}

func (impl *IHttpClient) SendPOSTJson(ctx context.Context, url string, body []byte, internalClientRes interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, impl.apiHost+url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := impl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//return xml.Unmarshal(data, internalClientRes)
	return json.Unmarshal(data, internalClientRes)
}

func (impl *IHttpClient) SendGETForm(ctx context.Context, url, params string, internalClientRes interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, impl.apiHost+url+"?"+params, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//	log.Info("SendGETForm: url:%s, req:%s", url, req)
	resp, err := impl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (impl *IHttpClient) SendPOSTForm(ctx context.Context, url string, body []byte, internalClientRes interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, impl.apiHost+url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF8")
	resp, err := impl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (impl *IHttpClient) SendGetHeader(ctx context.Context, url, headerKey, headerVal string, internalClientRes interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, impl.apiHost+url, nil)
	if err != nil {
		return err
	}
	req.Header.Set(headerKey, headerVal)
	resp, err := impl.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	if internalClientRes == nil {
		return nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, internalClientRes)
}

func (impl *IHttpClient) SendPOSTFormIgnoreErr(ctx context.Context, url string, body []byte, internalClientRes interface{}) (err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, impl.apiHost+url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := impl.client.Do(req)
	if err != nil {
		log.Error("impl.client.Do error:", err)
		if err.Error() == "400 Bad Request" {
			err = nil
		} else {
			return
		}
	}

	defer resp.Body.Close()
	if internalClientRes == nil {
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("ioutil.ReadAll error:", err)
		return
	}
	err = json.Unmarshal(data, internalClientRes)
	if err != nil {
		log.Error("json.Unmarshal error:", err)
	}
	return
}
