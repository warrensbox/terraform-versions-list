package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	"github.com/go-openapi/strfmt"
	semver "github.com/hashicorp/go-version"
)

type Release struct {
	Builds []struct {
		Arch string `json:"arch"`
		OS   string `json:"os"`
		URL  string `json:"url"`
	} `json:"builds"`
	IsPrerelease     bool `json:"is_prerelease"`
	LocalCacheTag    string
	TimestampCreated strfmt.DateTime `json:"timestamp_created"`
	Version          *semver.Version `json:"version"`
}

// httpGet : generic http get client for the given url and query parameters.
func httpGet(url *url.URL, values url.Values) (*http.Response, error) {
	url.RawQuery = values.Encode()

	res, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("[Error] : Retrieving contents from url %s\n: %q", url, err)

	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[Error] : non-200 response code during request: %d: Http status: %s", res.StatusCode, res.Status)
	}
	return res, nil
}

// getReleases : subfunc for GetTFReleases, used in a loop to get all terraform releases given the hashicorp url
func getReleases(url *url.URL, values url.Values) ([]*Release, error) {
	var releases []*Release
	resp, errURL := httpGet(url, values)
	if errURL != nil {
		return nil, fmt.Errorf("[Error] : Getting url: %q", errURL)
	}

	defer resp.Body.Close()
	body := new(bytes.Buffer)
	if _, err := io.Copy(body, resp.Body); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body.Bytes(), &releases); err != nil {
		return nil, fmt.Errorf("%q: %s", err, body.String())
	}
	return releases, nil
}

//GetTFReleases :  Get all terraform releases given the hashicorp url
func GetTFReleases(mirrorURL string) ([]*Release, error) {
	limit := 20
	u, err := url.Parse(mirrorURL)
	if err != nil {
		return nil, fmt.Errorf("[Error] : parsing url: %q", err)
	}
	values := u.Query()
	values.Set("limit", strconv.Itoa(limit))
	releaseSet, err := getReleases(u, values)
	if err != nil {
		return nil, err
	}
	var releases []*Release
	releases = append(releases, releaseSet...)
	for len(releaseSet) == limit {
		values.Set("after", releaseSet[len(releaseSet)-1].TimestampCreated.String())
		releaseSet, err = getReleases(u, values)
		if err != nil {
			return nil, err
		}
		releases = append(releases, releaseSet...)
	}

	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Version.GreaterThan(releases[j].Version)
	})
	return releases, nil
}
