package arias

import (
	"context"
	"github.com/google/uuid"
)

type Task interface {
	GetId() uuid.UUID
	Perform() error
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
}

func NewDownloadTask(server *Server, req DownloadRequest) DownloadTask {
	return &downloadTask{
		id:  uuid.New(),
		ctx: context.Background(),

		server: server,
		req:    req,
	}
}

func (task *downloadTask) GetId() uuid.UUID {
	return task.id
}

func (task *downloadTask) Perform() (err error) {
	err = task.Download()
	err = task.Upload()

	return
}

func (task *downloadTask) Download() error {

}

func (task *downloadTask) Upload() error {

}
