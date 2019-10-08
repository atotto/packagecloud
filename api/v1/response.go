package packagecloud

import (
	"io/ioutil"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func processResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		b, _ := ioutil.ReadAll(resp.Body)
		return status.Error(codes.Unauthenticated, string(b))
	case http.StatusNotFound:
		b, _ := ioutil.ReadAll(resp.Body)
		return status.Error(codes.NotFound, string(b))
	case http.StatusUnprocessableEntity:
		b, _ := ioutil.ReadAll(resp.Body)
		return status.Error(codes.AlreadyExists, string(b))
	default:
		b, _ := ioutil.ReadAll(resp.Body)
		return status.Errorf(codes.Internal, "resp: %s, %q", resp.Status, b)
	}
}
