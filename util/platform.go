package util

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/run/v1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var (
	httpClient        HTTPClient
	serviceCallClient ServiceCallClient
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ServiceCallClient interface {
	Do(opts ...googleapi.CallOption) (*run.Service, error)
}

func init() {
	httpClient = http.DefaultClient
}

func FetchURLByServiceName(ctx context.Context, name, region, projectID string) (string, error) {
	c, err := run.NewService(ctx)
	if err != nil {
		return "", err
	}
	c.BasePath = fmt.Sprintf("https://%s-run.googleapis.com/", region)

	return fetchServiceURL(
		c.Namespaces.Services.Get(fmt.Sprintf("namespaces/%s/services/%s", projectID, name)))
}

func fetchServiceURL(serviceCallClient ServiceCallClient) (string, error) {
	service, err := serviceCallClient.Do()
	if err != nil {
		return "", err
	}

	return service.Status.Url, nil
}

// check first env value of GOOGLE_CLOUD_PROJECT for local debug
func FetchProjectID() (string, error) {
	projectID, isSet := os.LookupEnv("GOOGLE_CLOUD_PROJECT")
	if isSet {
		return projectID, nil
	}

	req, err := http.NewRequest(http.MethodGet,
		"http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata-Flavor", "Google")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func IsCloudRun() bool {
	// There is no obvious way to detect whether the app is running on Clodu Run,
	// so we speculate from env var which is automatically added by Cloud Run.
	// ref. https://cloud.google.com/run/docs/reference/container-contract#env-vars
	// Note: we can't use K_SERVICE or K_REVISION since both are also used in Cloud Functions.
	return os.Getenv("K_CONFIGURATION") != ""
}

func GetIDToken(addr string) (string, error) {
	serviceURL := fmt.Sprintf("https://%s", strings.Split(addr, ":")[0])
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)

	return metadata.Get(tokenURL)
}

func FetchSecretLatestVersion(ctx context.Context, name, projectID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %v", err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, name),
	}

	resp, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(resp.Payload.Data), nil
}
