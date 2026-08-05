package storage

import "context"

// Storage abstracts the cloud object storage provider (e.g. Supabase, Firebase).
type Storage interface {
	Upload(ctx context.Context, path string, data []byte, contentType string) (url string, err error)
	Delete(ctx context.Context, path string) error
}

// Config holds the credentials needed to talk to the storage provider.
type Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
}

type client struct {
	cfg Config
}

// New creates a Storage client from the given config.
func New(cfg Config) Storage {
	return &client{cfg: cfg}
}

func (c *client) Upload(ctx context.Context, path string, data []byte, contentType string) (string, error) {
	// TODO: implement Supabase/Firebase upload
	return "", nil
}

func (c *client) Delete(ctx context.Context, path string) error {
	// TODO: implement Supabase/Firebase delete
	return nil
}
