package packagecloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/peterhellberg/link"
)

type CountValue struct {
	Value int `json:"value"`
}

type PackageDownloads struct {
	DownloadedAt time.Time `json:"downloaded_at"`
	IpAddress    string    `json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	Source       string    `json:"source"`
	ReadToken    string    `json:"read_token"`
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

func GetDownloadDetail(ctx context.Context, pkg PackageDetail, startDate, endDate string) ([]PackageDownloads, error) {
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
		Path:     pkg.DownloadsDetailURL,
		RawQuery: q.Encode(),
	}
	reqURL := u.String()
	var webLink map[string]*link.Link
	var details []PackageDownloads

	var next = &link.Link{}
	for ; next != nil; next = webLink["next"] {
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return nil, fmt.Errorf("http request: %s", err)
		}
		req.Header.Set("Accept", "application/json")
		token := packagecloudToken(ctx)
		req.SetBasicAuth(token, "")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http get: %s", err)
		}
		defer resp.Body.Close()

		total := resp.Header.Get("Total")
		perPage := resp.Header.Get("Per-Page")
		totalInt, _ := strconv.Atoi(total)
		perPageInt, _ := strconv.Atoi(perPage)

		if total != "" && perPage != "" && totalInt > perPageInt {
			webLink = link.ParseResponse(resp)
			if n, ok := webLink["next"]; ok {
				reqURL = n.URI
			}

		} else {
			next = nil
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var detail []PackageDownloads
			if err := json.NewDecoder(resp.Body).Decode(&detail); err != nil {
				return nil, fmt.Errorf("json decode: %s", err)
			}
			details = append(details, detail...)
		default:
			b, _ := ioutil.ReadAll(resp.Body)
			return nil, fmt.Errorf("resp: %s, %q", resp.Status, b)
		}

	}
	return details, nil
}
