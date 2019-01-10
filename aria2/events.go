package aria2

type DownloadEvent struct {
	GID string
}

func (event *DownloadEvent) String() string {
	return event.GID
}
