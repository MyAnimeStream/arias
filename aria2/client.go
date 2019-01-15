// aria2 is a go library to communicate with the aria2 rpc interface.
//
// aria2 is a utility for downloading files.
// The supported protocols are HTTP(S), FTP, SFTP, BitTorrent, and Metalink.
// aria2 can download a file from multiple sources/protocols and tries to utilize your maximum download bandwidth.
// It supports downloading a file from HTTP(S)/FTP /SFTP and BitTorrent at the same time,
// while the data downloaded from HTTP(S)/FTP/SFTP is uploaded to the BitTorrent swarm.
// Using Metalink chunk checksums, aria2 automatically validates chunks of data while downloading a file.
package aria2

import (
	"context"
	"errors"
	"fmt"
	"github.com/MyAnimeStream/arias/aria2/rpc"
	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
)

// URIs creates a string slice from the given uris
func URIs(uris ...string) []string {
	return uris
}

type EventListener func(event *DownloadEvent)

type Client struct {
	ws        *websocket.Conn
	rpcClient *rpc2.Client

	closed     bool
	listeners  map[string][]EventListener
	activeGIDs map[string]chan error
}

// Dial creates a new connection to an aria2 rpc interface.
// It returns a new client.
func Dial(url string) (client Client, err error) {
	dialer := websocket.Dialer{}

	ws, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		return
	}

	rwc := rpc.NewReadWriteCloser(ws)
	codec := jsonrpc.NewJSONCodec(&rwc)
	rpcClient := rpc2.NewClientWithCodec(codec)

	client = Client{ws: ws, rpcClient: rpcClient,
		closed:     false,
		listeners:  make(map[string][]EventListener),
		activeGIDs: make(map[string]chan error),
	}

	rpcClient.Handle("aria2.onDownloadStart", client.onDownloadStart)
	rpcClient.Handle("aria2.onDownloadPause", client.onDownloadPause)
	rpcClient.Handle("aria2.onDownloadStop", client.onDownloadStop)
	rpcClient.Handle("aria2.onDownloadComplete", client.onDownloadComplete)
	rpcClient.Handle("aria2.onDownloadError", client.onDownloadError)
	rpcClient.Handle("aria2.onBtDownloadComplete", client.onBtDownloadComplete)

	go rpcClient.Run()

	return
}

// Close closes the connection to the aria2 rpc interface.
// The client becomes unusable after that point.
func (c *Client) Close() error {
	c.closed = true

	err := c.rpcClient.Close()
	wsErr := c.ws.Close()
	if err == nil {
		err = wsErr
	}

	return err
}

func (c *Client) String() string {
	return fmt.Sprintf("Aria2Client")
}

func (c *Client) onEvent(name string, event *DownloadEvent) {
	listeners, ok := c.listeners[name]
	if !ok {
		return
	}

	for _, listener := range listeners {
		go listener(event)
	}
}

func (c *Client) onDownloadStart(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("downloadStart", event)
	return nil
}
func (c *Client) onDownloadPause(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("downloadPause", event)
	return nil
}
func (c *Client) onDownloadStop(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("downloadStop", event)
	channel, ok := c.activeGIDs[event.GID]
	if ok {
		channel <- errors.New("download stopped")
	}
	return nil
}
func (c *Client) onDownloadComplete(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("downloadComplete", event)
	channel, ok := c.activeGIDs[event.GID]
	if ok {
		channel <- nil
	}

	return nil
}
func (c *Client) onDownloadError(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("downloadError", event)
	channel, ok := c.activeGIDs[event.GID]
	if ok {
		channel <- errors.New("download encountered error")
	}
	return nil
}
func (c *Client) onBtDownloadComplete(_ *rpc2.Client, event *DownloadEvent, _ *interface{}) error {
	c.onEvent("btDownloadComplete", event)
	return nil
}

// Subscribe registers the given listener for an event.
// The listener will be called every time the event occurs.
func (c *Client) Subscribe(name string, listener EventListener) {
	listeners, ok := c.listeners[name]
	if !ok {
		listeners = make([]EventListener, 1)
		c.listeners[name] = listeners
	}

	c.listeners[name] = append(listeners, listener)
}

// WaitForDownload waits for a download denoted by its gid to finish.
func (c *Client) WaitForDownload(gid string) error {
	channel, ok := c.activeGIDs[gid]
	if !ok {
		channel = make(chan error, 1)
		c.activeGIDs[gid] = channel
	}

	err := <-channel
	delete(c.activeGIDs, gid)
	return err
}

// Download adds a new download and waits for it to complete.
// It returns the status of the finished download.
func (c *Client) Download(uris []string, options *Options) (status Status, err error) {
	return c.DownloadWithContext(context.Background(), uris, options)
}

// DownloadWithContext adds a new download and waits for it to complete.
// The passed context can be used to cancel the download.
// It returns the status of the finished download.
func (c *Client) DownloadWithContext(ctx context.Context, uris []string, options *Options) (status Status, err error) {
	gid, err := c.AddUri(uris, options)
	if err != nil {
		return
	}

	downloadDone := make(chan error, 1)

	go func() {
		downloadDone <- gid.WaitForDownload()
	}()

	select {
	case <-downloadDone:
		status, err = gid.TellStatus()
		if err != nil {
			return
		}
	case <-ctx.Done():
		_ = gid.Delete()
		err = errors.New("download cancelled")
	}

	return
}

// Delete removes the download denoted by gid and deletes all corresponding files.
func (c *Client) Delete(gid string) (err error) {
	err = c.Remove(gid)
	if err != nil {
		return
	}

	files, err := c.GetFiles(gid)
	if err == nil {
		for _, file := range files {
			_ = os.Remove(file.Path)
		}
	}

	return
}

// GetGID creates a GID struct which you can use to interact with the download directly
func (c *Client) GetGID(gid string) GID {
	return GID{c, gid}
}

// AddUri adds a new download.
// uris is a slice of HTTP/FTP/SFTP/BitTorrent URIs (strings) pointing to the same resource.
// If you mix URIs pointing to different resources,
// then the download may fail or be corrupted without aria2 complaining.
//
// When adding BitTorrent Magnet URIs, uris must have only one element and it should be BitTorrent Magnet URI.
//
// The new download is appended to the end of the queue.
// This method returns the GID of the newly registered download.
func (c *Client) AddUri(uris []string, options *Options) (GID, error) {
	args := []interface{}{uris}
	if options != nil {
		args = append(args, options)
	}

	var reply string
	err := c.rpcClient.Call("aria2.addUri", args, &reply)

	return c.GetGID(reply), err
}

// Remove removes the download denoted by gid.
// If the specified download is in progress, it is first stopped.
// The status of the removed download becomes removed.
func (c *Client) Remove(gid string) error {
	return c.rpcClient.Call("aria2.remove", []interface{}{gid}, nil)
}

// ForceRemove removes the download denoted by gid.
// This method behaves just like Remove() except that this method removes the download
// without performing any actions which take time, such as contacting BitTorrent trackers to
// unregister the download first.
func (c *Client) ForceRemove(gid string) error {
	return c.rpcClient.Call("aria2.forceRemove", []interface{}{gid}, nil)
}

// Pause pauses the download denoted by gid.
// The status of paused download becomes paused. If the download was active,
// the download is placed in the front of the queue. While the status is paused,
// the download is not started. To change status to waiting, use the Unpause() method.
func (c *Client) Pause(gid string) error {
	return c.rpcClient.Call("aria2.pause", []interface{}{gid}, nil)
}

// PauseAll is equal to calling Pause() for every active/waiting download.
func (c *Client) PauseAll() error {
	return c.rpcClient.Call("aria2.pauseAll", nil, nil)
}

// ForcePause pauses the download denoted by gid.
// This method behaves just like Pause() except that this method pauses downloads
// without performing any actions which take time, such as contacting BitTorrent trackers to
// unregister the download first.
func (c *Client) ForcePause(gid string) error {
	return c.rpcClient.Call("aria2.forcePause", []interface{}{gid}, nil)
}

// ForcePauseAll is equal to calling ForcePause() for every active/waiting download.
func (c *Client) ForcePauseAll() error {
	return c.rpcClient.Call("aria2.forcePauseAll", nil, nil)
}

// Unpause changes the status of the download denoted by gid from paused to waiting,
// making the download eligible to be restarted.
func (c *Client) Unpause(gid string) error {
	return c.rpcClient.Call("aria2.unpause", []interface{}{gid}, nil)
}

// UnpauseAll is equal to calling Unpause() for every paused download.
func (c *Client) UnpauseAll() error {
	return c.rpcClient.Call("aria2.unpauseAll", nil, nil)
}

// TellStatus returns the progress of the download denoted by gid.
// If specified, the response only contains only the keys passed to the method.
// If keys is empty, the response contains all keys.
// This is useful when you just want specific keys and avoid unnecessary transfers.
func (c *Client) TellStatus(gid string, keys ...string) (Status, error) {
	args := []interface{}{gid}
	if len(keys) > 0 {
		args = append(args, keys)
	}

	var reply Status
	err := c.rpcClient.Call("aria2.tellStatus", args, &reply)

	return reply, err
}

// GetURIs returns the URIs used in the download denoted by gid.
// The response is a slice of URI.
func (c *Client) GetURIs(gid string) ([]URI, error) {
	var reply []URI
	err := c.rpcClient.Call("aria2.getUris", []interface{}{gid}, &reply)

	return reply, err
}

// GetFiles returns the file list of the download denoted by gid.
// The response is a slice of File.
func (c *Client) GetFiles(gid string) ([]File, error) {
	var reply []File
	err := c.rpcClient.Call("aria2.getFiles", []interface{}{gid}, &reply)

	return reply, err
}

type PositionSetBehaviour string

const (
	SetPositionStart    PositionSetBehaviour = "POS_SET"
	SetPositionEnd                           = "POS_END"
	SetPositionRelative                      = "POS_CUR"
)

// ChangePosition changes the position of the download denoted by gid in the queue.
// If how is SetPositionStart, it moves the download to a position relative to the beginning of the queue.
// If how is SetPositionRelative, it moves the download to a position relative to the current position.
// If how is SetPositionEnd, it moves the download to a position relative to the end of the queue.
// If the destination position is less than 0 or beyond the end of the queue,
// it moves the download to the beginning or the end of the queue respectively.
// The response is an integer denoting the resulting position.
func (c *Client) ChangePosition(gid string, pos int, how PositionSetBehaviour) (int, error) {
	args := []interface{}{gid, pos}
	if how != "" {
		args = append(args, how)
	}

	var reply int
	err := c.rpcClient.Call("aria2.changePosition", args, &reply)

	return reply, err
}
