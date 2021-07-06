package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/run/v1"
)

type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func TestFetchProjectID(t *testing.T) {
	wantProjectID := "sample-project-id"

	httpClient = &MockHTTPClient{
		DoFunc: func(*http.Request) (*http.Response, error) {
			r := ioutil.NopCloser((bytes.NewReader([]byte(wantProjectID))))
			return &http.Response{
				StatusCode: 200,
				Body:       r,
			}, nil
		},
	}

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", wantProjectID); err != nil {
		t.Fatal(err)
	}

	projectID, err := FetchProjectID()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := wantProjectID, projectID; want != got {
		t.Errorf("wrong request %q, want %q", got, want)
	}

	if err := os.Unsetenv("GOOGLE_CLOUD_PROJECT"); err != nil {
		t.Fatal(err)
	}

	projectID, err = FetchProjectID()
	if err != nil {
		t.Fatal(err)
	}

	if want, got := wantProjectID, projectID; want != got {
		t.Errorf("wrong response %s, want %s", got, want)
	}
}

type MockServiceCallClient struct {
	DoFunc func(opts ...googleapi.CallOption) (*run.Service, error)
}

func (m *MockServiceCallClient) Do(opts ...googleapi.CallOption) (*run.Service, error) {
	return m.DoFunc(opts...)
}

func TestFetchServiceURL(t *testing.T) {
	wantURL := "http://dummy.com"

	serviceCallClient = &MockServiceCallClient{
		DoFunc: func(opts ...googleapi.CallOption) (*run.Service, error) {
			return &run.Service{
				Status: &run.ServiceStatus{Url: wantURL},
			}, nil
		},
	}

	url, err := fetchServiceURL(serviceCallClient)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := wantURL, url; wantURL != url {
		t.Errorf("wrong response %s, want %s", got, want)
	}
}
