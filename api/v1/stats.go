package packagecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type CountValue struct {
	Value int `json:"value"`
}

func GetDownloadCount(ctx context.Context, pkg PackageDetail, startDate, endDate string) (*CountValue, error) {
	q := url.Values{}
	if startDate != "" {
		q.Add("start_date", startDate)
	}
	if endDate != "" {
		q.Add("end_date", endDate)
	}
	u := &url.URL{
		Scheme:   "https",
		Host:     "packagecloud.io",
		Path:     pkg.DownloadsCountURL,
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error http get: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error http request: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var count CountValue
		if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
			return nil, fmt.Errorf("json decode: %v", err)
		}
		return &count, nil
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("resp_status: %s, %q", resp.Status, b)
	}
}
