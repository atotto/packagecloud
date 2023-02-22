package packagecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

var distributions = &Distributions{}

var cacheDistributionsOnce sync.Once

type Distributions struct {
	Deb     []Distribution `json:"deb"`
	Rpm     []Distribution `json:"rpm"`
	Py      []Distribution `json:"py"`
	Jar     []Distribution `json:"jar"`
	Node    []Distribution `json:"node"`
	Alpine  []Distribution `json:"alpine"`
	Anyfile []Distribution `json:"anyfile"`
	Helm    []Distribution `json:"helm"`
	Dsc     []Distribution `json:"dsc"`
}

func GetDistributions(ctx context.Context) (*Distributions, error) {
	if err := cacheDistributions(ctx); err != nil {
		return nil, err
	}
	return distributions, nil
}

func cacheDistributions(ctx context.Context) (err error) {
	cacheDistributionsOnce.Do(func() {
		distributions, err = getDistributions(ctx)
	})
	return err
}

func getDistributions(ctx context.Context) (*Distributions, error) {
	req, err := http.NewRequest("GET", "https://packagecloud.io/api/v1/distributions.json", nil)
	if err != nil {
		return nil, fmt.Errorf("http request: %s", err)
	}
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %s", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var d Distributions
		if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
			return nil, fmt.Errorf("json decode: %s", err)
		}
		return &d, nil
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("resp: %s, %q", resp.Status, b)
	}
}

func (d *Distributions) DebianDistroVersionID(distro, version string) (id string, ok bool) {
	for _, di := range distributions.Deb {
		if di.IndexName == distro {
			for _, ver := range di.Versions {
				if ver.IndexName == version {
					return strconv.Itoa(ver.ID), true
				}
			}
		}
	}
	return "", false
}

func findDistroVersionID(distributions []Distribution, distro, version string) (id string, ok bool) {
	for _, di := range distributions {
		if di.IndexName == distro {
			for _, ver := range di.Versions {
				if ver.IndexName == version {
					return strconv.Itoa(ver.ID), true
				}
			}
		}
	}
	return "", false
}

type Versions struct {
	ID            int    `json:"id"`
	DisplayName   string `json:"display_name"`
	IndexName     string `json:"index_name"`
	VersionNumber string `json:"version_number"`
}

type Distribution struct {
	DisplayName string     `json:"display_name"`
	IndexName   string     `json:"index_name"`
	Versions    []Versions `json:"versions"`
}
