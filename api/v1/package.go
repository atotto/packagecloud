package packagecloud

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PushPackage(ctx context.Context, repos, distro, version string, fpath string) error {
	distroVersionID, ok := distributions.DistroVersionID(distro, version)
	if !ok {
		return status.Errorf(codes.InvalidArgument, "unknown distribution: %s/%s", distro, version)
	}

	var r io.ReadCloser
	var err error
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
	if err := mw.WriteField(`package[distro_version_id]`, distroVersionID); err != nil {
		return status.Errorf(codes.InvalidArgument, "multipart: %s", err)
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
