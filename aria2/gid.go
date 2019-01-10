package aria2

type GID struct {
	client *Client
	GID    string
}

func (gid *GID) String() string {
	return gid.GID
}

func (gid *GID) WaitForDownload() error {
	return gid.client.WaitForDownload(gid.GID)
}

func (gid *GID) Remove() error {
	_, err := gid.client.Remove(gid.GID)
	return err
}
func (gid *GID) ForceRemove() error {
	_, err := gid.client.ForceRemove(gid.GID)
	return err
}

func (gid *GID) Pause() error {
	_, err := gid.client.Pause(gid.GID)
	return err
}
func (gid *GID) ForcePause() error {
	_, err := gid.client.ForcePause(gid.GID)
	return err
}

func (gid *GID) Unpause() error {
	_, err := gid.client.Unpause(gid.GID)
	return err
}

func (gid *GID) TellStatus() (Status, error) {
	return gid.client.TellStatus(gid.GID)
}

func (gid *GID) GetURIs() ([]URI, error) {
	return gid.client.GetURIs(gid.GID)
}

func (gid *GID) GetFiles() ([]File, error) {
	return gid.client.GetFiles(gid.GID)
}
