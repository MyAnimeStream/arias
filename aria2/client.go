package aria2

import (
	"errors"
	"fmt"
	"github.com/cenkalti/rpc2"
	"github.com/cenkalti/rpc2/jsonrpc"
	"github.com/gorilla/websocket"
	"github.com/myanimestream/arias/aria2/rpc"
	"net/http"
)

type EventListener func(event *DownloadEvent)

type Client struct {
	ws        *websocket.Conn
	rpcClient *rpc2.Client

	closed     bool
	listeners  map[string][]EventListener
	activeGIDs map[string]chan error
}

func NewClient(url string) (*Client, error) {
	dialer := websocket.Dialer{}

	ws, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		return nil, err
	}

	rwc := rpc.NewReadWriteCloser(ws)
	codec := jsonrpc.NewJSONCodec(&rwc)
	rpcClient := rpc2.NewClientWithCodec(codec)

	client := Client{ws: ws, rpcClient: rpcClient,
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

	return &client, nil
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

func (c *Client) Close() error {
	c.closed = true

	err := c.rpcClient.Close()
	wsErr := c.ws.Close()
	if err == nil {
		err = wsErr
	}

	return err
}

func (c *Client) WaitForDownload(gid string) error {
	channel, ok := c.activeGIDs[gid]
	if !ok {
		channel = make(chan error, 1)
		c.activeGIDs[gid] = channel
	}

	return <-channel
}

func (c *Client) Subscribe(name string, listener EventListener) {
	listeners, ok := c.listeners[name]
	if !ok {
		listeners = make([]EventListener, 1)
		c.listeners[name] = listeners
	}

	c.listeners[name] = append(listeners, listener)
}

func (c *Client) AddUri(uris []string, options *Options) (GID, error) {
	args := []interface{}{uris}
	if options != nil {
		args = append(args, options)
	}

	var reply string
	err := c.rpcClient.Call("aria2.addUri", args, &reply)

	gid := GID{c, reply}
	return gid, err
}

func (c *Client) Remove(gid string) (string, error) {
	var reply string
	err := c.rpcClient.Call("aria2.remove", []interface{}{gid}, &reply)

	return reply, err
}

func (c *Client) ForceRemove(gid string) (string, error) {
	var reply string
	err := c.rpcClient.Call("aria2.forceRemove", []interface{}{gid}, &reply)

	return reply, err
}

func (c *Client) Pause(gid string) (string, error) {
	var reply string
	err := c.rpcClient.Call("aria2.pause", []interface{}{gid}, &reply)

	return reply, err
}

func (c *Client) PauseAll() error {
	return c.rpcClient.Call("aria2.pauseAll", nil, nil)
}

func (c *Client) ForcePause(gid string) (string, error) {
	var reply string
	err := c.rpcClient.Call("aria2.forcePause", []interface{}{gid}, &reply)

	return reply, err
}

func (c *Client) ForcePauseAll() error {
	return c.rpcClient.Call("aria2.forcePauseAll", nil, nil)
}

func (c *Client) Unpause(gid string) (string, error) {
	var reply string
	err := c.rpcClient.Call("aria2.unpause", []interface{}{gid}, &reply)

	return reply, err
}

func (c *Client) UnpauseAll() error {
	return c.rpcClient.Call("aria2.unpauseAll", nil, nil)
}

func (c *Client) TellStatus(gid string, keys ...string) (Status, error) {
	args := []interface{}{gid}
	if len(keys) > 0 {
		args = append(args, keys)
	}

	var reply Status
	err := c.rpcClient.Call("aria2.tellStatus", args, &reply)

	return reply, err
}

func (c *Client) GetURIs(gid string) ([]URI, error) {
	var reply []URI
	err := c.rpcClient.Call("aria2.getUris", []interface{}{gid}, &reply)

	return reply, err
}

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

func (c *Client) ChangePosition(gid string, pos int, how PositionSetBehaviour) (int, error) {
	args := []interface{}{gid, pos}
	if how != "" {
		args = append(args, how)
	}

	var reply int
	err := c.rpcClient.Call("aria2.changePosition", args, &reply)

	return reply, err
}
