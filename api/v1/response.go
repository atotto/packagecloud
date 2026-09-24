package packagecloud

import (
	"io"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func processResponse(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		b, _ := io.ReadAll(resp.Body)
		return status.Error(codes.Unauthenticated, string(b))
	case http.StatusNotFound:
		b, _ := io.ReadAll(resp.Body)
		return status.Error(codes.NotFound, string(b))
	case http.StatusUnprocessableEntity:
		b, _ := io.ReadAll(resp.Body)
		return status.Error(codes.AlreadyExists, string(b))
	default:
		b, _ := io.ReadAll(resp.Body)
		return status.Errorf(codes.Internal, "resp: %s, %q", resp.Status, b)
	}
}
