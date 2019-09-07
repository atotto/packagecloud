package packagecloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type client struct {
}

func NewPackagecloud(ctx context.Context) {

}

func PushPackage(ctx context.Context, repos, distro, version string, fpath string) error {
	distroVersionID, ok := distributions.DistroVersionID(distro, version)
	if !ok {
		return fmt.Errorf("unknown distribution: %s/%s", distro, version)
	}

	f, err := os.Open(fpath)
	if err != nil {
		return fmt.Errorf("file open: %s", err)
	}
	defer f.Close()
	_, fname := filepath.Split(fpath)

	form := url.Values{}
	form.Add("package[distro_version_id]", distroVersionID)
	form.Add("package[distro_version_id]", distroVersionID)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField(`package[distro_version_id]`, distroVersionID); err != nil {
		return fmt.Errorf("multipart: %s", err)
	}
	w, err := mw.CreateFormFile(`package[package_file]`, fname)
	if err != nil {
		return fmt.Errorf("multipart: %s", err)
	}
	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("file read: %s", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("multipart close: %s", err)
	}

	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/packages.json", repos)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return fmt.Errorf("http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %s", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("already pushed")
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("resp: %s, %q", resp.Status, b)
	}

	return nil
}

func PromotePackage(ctx context.Context, dstRepos, srcRepo, distro, version string, fpath string) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField(`destination`, dstRepos); err != nil {
		return fmt.Errorf("multipart: %s", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("multipart close: %s", err)
	}

	_, fname := filepath.Split(fpath)
	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/%s/%s/%s/promote.json", srcRepo, distro, version, fname)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %s", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("not found")
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("resp: %s, %q", resp.Status, b)
	}

	return nil
}

func DeletePackage(ctx context.Context, repos, distro, version string, fpath string) error {
	_, fname := filepath.Split(fpath)
	url := fmt.Sprintf("https://packagecloud.io/api/v1/repos/%s/%s/%s/%s", repos, distro, version, fname)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("http request: %s", err)
	}
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Accept", "application/json")

	token := packagecloudToken(ctx)
	req.SetBasicAuth(token, "")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %s", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("not found")
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("resp: %s, %q", resp.Status, b)
	}

	return nil
}
