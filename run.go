package sdk

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"google.golang.org/api/run/v1"
)

func FetchURLByServiceName(ctx context.Context, name, region string) (string, error) {
	c, err := run.NewService(ctx)
	if err != nil {
		return "", err
	}
	c.BasePath = fmt.Sprintf("https://%s-run.googleapis.com/", region)

	projectID, err := FetchProjectID()
	if err != nil {
		return "", err
	}

	service, err := c.Namespaces.Services.Get(fmt.Sprintf("namespaces/%s/services/%s", projectID, name)).Do()
	if err != nil {
		return "", err
	}

	return service.Status.Url, nil
}

func FetchProjectID() (string, error) {
	if !IsCloudRun() {
		return os.Getenv("GOOGLE_CLOUD_PROJECT"), nil
	}
	req, err := http.NewRequest("GET",
		"http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := http.DefaultClient.Do(req)
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
