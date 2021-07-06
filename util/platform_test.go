package util

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var (
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func TestMain(m *testing.M) {
	Client = &MockClient{}
	os.Exit(m.Run())
}

func TestFetchProjectID(t *testing.T) {
	wantProjectID := "sample-project-id"

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		r := ioutil.NopCloser((bytes.NewReader([]byte(wantProjectID))))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
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
		t.Errorf("wrong request %q, want %q", got, want)
	}
}
