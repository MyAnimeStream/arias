package arias

import (
	"cloud.google.com/go/storage"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"google.golang.org/api/option"
	"io"
	"net/http"
)

type UploadOptions struct {
	Bucket      string
	Filename    string
	ContentType string
	// ForceGZip
	ForceGZip bool
}

type UploadOutput struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
}

type Storage interface {
	Upload(ctx context.Context, f io.ReadSeeker, options UploadOptions) (UploadOutput, error)
}

func NewStorageFromType(storageType string) (s Storage, err error) {
	switch storageType {
	case "google":
		s, err = NewGoogleCloudStorage()
	case "s3":
		s, err = NewS3Storage()
	default:
		err = fmt.Errorf("unknown storage: %s", storageType)
	}

	return
}

func determineContentType(f io.ReadSeeker, options UploadOptions) (contentType string, err error) {
	contentType = options.ContentType

	if contentType == "" {
		buffer := make([]byte, 512)
		_, err = f.Read(buffer)
		if err != nil {
			return
		}

		_, err = f.Seek(0, 0)
		if err != nil {
			return
		}

		contentType = http.DetectContentType(buffer)
	}

	return
}

type googleCloudStorage struct {
	ctx    context.Context
	client *storage.Client
}

func NewGoogleCloudStorage(opts ...option.ClientOption) (s Storage, err error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return
	}

	s = &googleCloudStorage{
		ctx:    ctx,
		client: client,
	}

	return
}

func (s *googleCloudStorage) Upload(ctx context.Context, f io.ReadSeeker, options UploadOptions) (UploadOutput, error) {
	var out UploadOutput

	bkt := s.client.Bucket(options.Bucket)
	obj := bkt.Object(options.Filename)
	objWriter := obj.NewWriter(ctx)
	attrs := objWriter.ObjectAttrs

	contentType, err := determineContentType(f, options)
	if err != nil {
		return out, nil
	}

	attrs.ContentType = contentType
	attrs.ContentEncoding = "gzip"

	// Google supports gzip "natively". It automatically decodes the data if need be
	w, _ := gzip.NewWriterLevel(objWriter, gzip.BestCompression)
	_, errWrite := io.Copy(w, f)
	_ = w.Close()
	_ = objWriter.Close()

	if errWrite != nil {
		return out, errWrite
	}

	objAttrs, err := obj.Attrs(ctx)
	if err == nil {
		out.Bucket = objAttrs.Bucket
		out.Filename = objAttrs.Name
	}

	return out, nil
}

type s3Storage struct {
	session  *session.Session
	uploader *s3manager.Uploader
}

func NewS3Storage(opts ...*aws.Config) (s Storage, err error) {
	opt := aws.NewConfig().WithCredentials(credentials.NewEnvCredentials())
	opts = append([]*aws.Config{opt}, opts...)
	sess, err := session.NewSession(opts...)
	if err != nil {
		return
	}

	sess.Config.WithCredentialsChainVerboseErrors(true)
	uploader := s3manager.NewUploader(sess)

	s = &s3Storage{
		session:  sess,
		uploader: uploader,
	}

	return
}

func (s *s3Storage) Upload(ctx context.Context, f io.ReadSeeker, options UploadOptions) (out UploadOutput, err error) {
	contentType, err := determineContentType(f, options)
	if err != nil {
		return
	}

	var reader io.ReadSeeker
	var contentEncoding *string
	if options.ForceGZip {
		contentEncoding = aws.String("gzip")
		reader, writer := io.Pipe()

		defer func() { _ = reader.Close() }()

		go func() {
			w, _ := gzip.NewWriterLevel(writer, gzip.BestCompression)
			_, _ = io.Copy(w, f)
			_ = w.Close()
			_ = writer.Close()
		}()
	} else {
		reader = f
	}

	input := s3manager.UploadInput{
		Bucket:          &options.Bucket,
		Key:             &options.Filename,
		ContentType:     &contentType,
		ContentEncoding: contentEncoding,
		Body:            reader,
	}
	_, err = s.uploader.UploadWithContext(ctx, &input)

	out = UploadOutput{options.Bucket, options.Filename}

	return
}
