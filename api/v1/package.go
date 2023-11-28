package packagecloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/peterhellberg/link"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PushPackage(ctx context.Context, repos, distro, version string, fpath string) error {
	var distroVersionID string
	var ok bool

	ds, err := GetDistributions(ctx)
	if err != nil {
		return err
	}
	switch filepath.Ext(fpath) {
	case ".deb":
		distroVersionID, ok = findDistroVersionID(ds.Deb, distro, version)
	case ".dsc":
		distroVersionID, ok = findDistroVersionID(ds.Dsc, distro, version)
	case ".rpm":
		distroVersionID, ok = findDistroVersionID(ds.Rpm, distro, version)
	case ".apk":
		distroVersionID, ok = findDistroVersionID(ds.Alpine, distro, version)
	case ".whl":
		distro = "python"
		version = ""
		distroVersionID, ok = findDistroVersionID(ds.Py, distro, version)
	}
	if !ok {
		return status.Errorf(codes.InvalidArgument, "unknown distribution: %s/%s", distro, version)
	}

	var r io.ReadCloser
	if strings.HasPrefix(fpath, "http://") || strings.HasPrefix(fpath, "https://") {
		resp, err := http.Get(fpath)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "http GET: %s", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode > 400 {
			body, _ := ioutil.ReadAll(resp.Body)
			return status.Errorf(codes.InvalidArgument, "http GET: %s\n>> %q", resp.Status, body)
		}
		r = resp.Body
	} else {
		r, err = os.Open(fpath)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "file open: %s", err)
		}
		defer r.Close()
	}
	_, fname := filepath.Split(fpath)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if distroVersionID != "" {
		if err := mw.WriteField(`package[distro_version_id]`, distroVersionID); err != nil {
			return status.Errorf(codes.InvalidArgument, "multipart: %s", err)
		}
	}
	w, err := mw.CreateFormFile(`package[package_file]`, fname)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "multipart: %s", err)
	}
	if _, err := io.Copy(w, r); err != nil {
		return status.Errorf(codes.InvalidArgument, "file read: %s", err)
	}
	if err := mw.Close(); err != nil {
		return status.Errorf(codes.InvalidArgument, "multipart close: %s", err)
	}

	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/packages.json", repos)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http post: %s", err)
	}
	defer resp.Body.Close()

	return processResponse(resp)
}

type PackageDetail struct {
	Name               string    `json:"name"`
	Arch               string    `json:"architecture"`
	Release            string    `json:"release"`
	DistroVersion      string    `json:"distro_version"`
	CreateTime         time.Time `json:"created_at"`
	Version            string    `json:"version"`
	Type               string    `json:"type"`
	Filename           string    `json:"filename"`
	UploaderName       string    `json:"uploader_name"`
	Indexed            bool      `json:"indexed"`
	PackageURL         string    `json:"package_url"`
	DownloadURL        string    `json:"download_url"`
	DownloadsCountURL  string    `json:"downloads_count_url"`
	DownloadsDetailURL string    `json:"downloads_detail_url"`
}

func SearchPackage(ctx context.Context, repos, distro string, perPage int, query, filter string) ([]PackageDetail, error) {
	q := url.Values{}
	if distro != "" {
		q.Add("dist", distro)
	}
	if query != "" {
		q.Add("q", query)
	}
	if perPage != 0 {
		q.Add("per_page", strconv.Itoa(perPage))
	}

	if filter != "" {
		q.Add("filter", filter)
	}

	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/search?%s", repos, q.Encode())
	var webLink map[string]*link.Link
	var details []PackageDetail

	var next = &link.Link{}
	for ; next != nil; next = webLink["next"] {
		req, err := http.NewRequest("GET", url, nil)
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

		total := resp.Header.Get("Total")
		perPage := resp.Header.Get("Per-Page")
		totalInt, _ := strconv.Atoi(total)
		perPageInt, _ := strconv.Atoi(perPage)

		if total != "" && perPage != "" && totalInt > perPageInt {
			webLink = link.ParseResponse(resp)
			if n, ok := webLink["next"]; ok {
				url = n.URI
			}

		} else {
			next = nil
		}

		switch resp.StatusCode {
		case http.StatusOK:
			var detail []PackageDetail
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

func PromotePackage(ctx context.Context, dstRepos, srcRepo, distro, version string, fpath string) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField(`destination`, dstRepos); err != nil {
		return status.Errorf(codes.InvalidArgument, "multipart: %s", err)
	}
	if err := mw.Close(); err != nil {
		return status.Errorf(codes.InvalidArgument, "multipart close: %s", err)
	}

	_, fname := filepath.Split(fpath)
	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/%s/%s/%s/promote.json", srcRepo, distro, version, fname)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http post: %s", err)
	}
	defer resp.Body.Close()

	return processResponse(resp)
}

func DeletePackage(ctx context.Context, repos, distro, version string, fpath string) error {
	_, fname := filepath.Split(fpath)
	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/%s/%s/%s", repos, distro, version, fname)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "http post: %s", err)
	}
	defer resp.Body.Close()

	return processResponse(resp)
}
