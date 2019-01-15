package arias

import "errors"

type DownloadRequest struct {
	Url    string `schema:"url"`
	Bucket string `schema:"bucket"`
	Name   string `schema:"name"`

	CallbackUrl string `schema:"callback"`
}

func (req *DownloadRequest) UseConfig(c *Config) error {
	if req.Bucket == "" {
		req.Bucket = c.DefaultBucket
	} else if !c.AllowBucketOverride {
		return errors.New("bucket override forbidden")
	}

	if req.Name == "" && !c.AllowNoName {
		return errors.New("name must be provided")
	}

	return nil
}

func (req *DownloadRequest) Check() error {
	switch {
	case req == nil:
		return errors.New("request is empty")
	case req.Url == "":
		return errors.New("url not specified")
	}

	return nil
}

type DownloadResponse struct {
	Id string `json:"id"`
}
