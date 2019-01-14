package arias

import "errors"

type DownloadRequest struct {
	Url  string `schema:"url"`
	Name string `schema:"name"`
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
	ID string `json:"id"`
}
