package s3

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/c032/go-objectstorage"
)

var _ objectstorage.Store = (*S3)(nil)

type S3 struct {
	bucket  string
	session client.ConfigProvider
}

func (s *S3) Read(key string, w io.Writer) (int64, error) {
	downloader := s3manager.NewDownloader(s.session)

	downloader.Concurrency = 1

	sw := &sequentialWriter{
		w: w,
	}

	var (
		err error

		n int64
	)
	n, err = downloader.Download(sw, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return n, fmt.Errorf("could not download file: %w", err)
	}

	return n, nil
}

func (s *S3) Put(key string, r io.Reader) error {
	uploader := s3manager.NewUploader(s.session)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("could not upload file: %w", err)
	}

	return nil
}

func (s *S3) Remove(key string) error {
	svc := s3.New(s.session)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := svc.DeleteObject(input)
	if err != nil {
		return fmt.Errorf("could not delete: %w", err)
	}

	return nil
}

func NewFromSession(c client.ConfigProvider, bucket string) (*S3, error) {
	s := &S3{
		bucket:  bucket,
		session: c,
	}

	return s, nil
}
