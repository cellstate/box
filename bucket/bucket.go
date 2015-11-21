package bucket

type Bucket interface {
	Ping() error //returns an error when the bucket is not online
}

func Create(uri string) (Bucket, error) {

	//@todo in the future, support multiple backends
	return NewS3(uri)
}
