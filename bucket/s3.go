package bucket

import (
	"net/url"

	"github.com/cellstate/box/config"
)

type S3 struct {
	endpoint *url.URL
}

func NewS3(uri string) (*S3, error) {
	loc, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	return &S3{
		endpoint: loc,
	}, nil
}

func (s *S3) Ping() error {
	return nil
}

func (s *S3) Config() *config.BucketConfig {
	return &config.BucketConfig{
		Endpoint: s.endpoint.String(),
	}
}
