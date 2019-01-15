package arias

import (
	"context"
	"errors"
	"fmt"
	"github.com/MyAnimeStream/arias/aria2"
	"github.com/google/uuid"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Task interface {
	GetId() uuid.UUID
	GetStatus() *TaskStatus
	Perform() error
}

type TaskStatus struct {
	Running bool        `json:"running"`
	State   string      `json:"state"`
	Result  interface{} `json:"result,omitempty"`
	Err     interface{} `json:"error,omitempty"`
}

func NewTaskStatus() *TaskStatus {
	return &TaskStatus{State: "waiting"}
}

func (status *TaskStatus) Start() {
	status.Running = true
	status.State = "started"
}

func (status *TaskStatus) EnterState(state string) {
	status.State = state
}

func (status *TaskStatus) Error(err error) {
	status.Running = false
	status.Err = err.Error()
}

func (status *TaskStatus) Done(res interface{}) {
	status.Running = false
	status.State = "done"
	status.Result = res
}

func formatFilename(template string, filename string) string {
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	r := strings.NewReplacer("{filename}", filename, "{name}", name, "{ext}", ext)
	return r.Replace(template)
}

type DownloadTask interface {
	Task
	Download() error
	Upload() error
}

type downloadTask struct {
	id uuid.UUID

	ctx    context.Context
	server *Server

	req DownloadRequest

	status *TaskStatus
	gid    *aria2.GID
	file   *aria2.File
	result UploadOutput
}

func NewDownloadTask(server *Server, req DownloadRequest) DownloadTask {
	return &downloadTask{
		id:  uuid.New(),
		ctx: context.Background(),

		server: server,
		req:    req,
		status: NewTaskStatus(),
	}
}

func (task *downloadTask) GetId() uuid.UUID {
	return task.id
}

func (task *downloadTask) String() string {
	return fmt.Sprintf("Download task: %s\n", task.GetId())
}

func (task *downloadTask) Perform() (err error) {
	defer func() { _ = task.Cleanup() }()
	defer task.SendCallback()

	task.status.Start()

	log.Printf("[%s] download started\n", task.id)
	task.status.EnterState("downloading")
	err = task.Download()
	if err != nil {
		log.Printf("[%s] download failed: %s\n", task.id, err)
		task.status.Error(err)
		return
	}

	log.Printf("[%s] upload started\n", task.id)
	task.status.State = "uploading"
	err = task.Upload()
	if err != nil {
		log.Printf("[%s] upload failed: %s\n", task.id, err)
		task.status.Error(err)
		return
	}

	log.Printf("[%s] done\n", task.id)
	task.status.Done(task.result)
	return
}

func (task *downloadTask) SendCallback() {
	if task.req.CallbackUrl != "" {
		task.server.GoSendCallback(task.req.CallbackUrl, task.status)
	}
}

func (task *downloadTask) GetStatus() *TaskStatus {
	return task.status
}

func (task *downloadTask) Download() error {
	ariaClient := &task.server.AriaClient
	status, err := ariaClient.DownloadWithContext(task.ctx, aria2.URIs(task.req.Url), &aria2.Options{})
	gid := ariaClient.GetGID(status.GID)
	task.gid = &gid

	if err != nil {
		return err
	}

	files := status.Files
	if len(files) != 1 {
		return fmt.Errorf("invalid number of files downloaded: %d", len(files))
	}

	task.file = &files[0]
	return nil
}

func (task *downloadTask) Upload() error {
	file := task.file
	if file == nil {
		return errors.New("no file to upload")
	}

	name := task.req.Name
	if name != "" {
		name = formatFilename(name, file.Path)
	} else {
		name = path.Base(file.Path)
	}

	f, err := os.Open(file.Path)
	if err != nil {
		return err
	}

	storage := task.server.Storage
	result, err := storage.Upload(task.ctx, f, UploadOptions{Bucket: task.req.Bucket, Filename: name})
	_ = f.Close()

	task.result = result

	return err
}

func (task *downloadTask) Cleanup() (err error) {
	if task.gid != nil {
		err = task.gid.Delete()
	}

	return
}
