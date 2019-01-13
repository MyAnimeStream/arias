package arias

type DownloadRequest struct {
	Url  string `schema:"url"`
	Name string `schema:"name"`
}

func defaultDownloadRequest() DownloadRequest {
	return DownloadRequest{}
}
