package bucket

import (
	"github.com/cellstate/box/config"
)

type Bucket interface {
	Ping() error //returns an error when the bucket is not online
	Config() *config.BucketConfig
}

func Create(uri string) (Bucket, error) {

	//@todo in the future, support multiple backends
	return NewS3(uri)
}
