package cloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// AWSSession represent a live session to AWS services
type AWSSession struct {
	s      *session.Session
	bucket string
}

// Exists return true if a path does really exist
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !errors.Is(err, os.ErrNotExist)
	}
	return true
}

// NewAWSSessionFromEnvironment create a new session from environment variable
func NewAWSSessionFromEnvironment() (*AWSSession, error) {
	return NewAWSSession("", "", os.Getenv("AWS_S3_ENDPOINT"), os.Getenv("AWS_S3_REGION"), os.Getenv("AWS_S3_BUCKET"))
}

// NewAWSSession create a new session
func NewAWSSession(akid string, secret string, endpoint string, region string, bucket string) (*AWSSession, error) {
	var cred *credentials.Credentials

	if akid != "" && secret != "" {
		cred = credentials.NewStaticCredentials(akid, secret, "")
	}

	s, err := session.NewSession(
		&aws.Config{
			Endpoint:    aws.String(endpoint),
			Region:      aws.String(region),
			Credentials: cred,
		},
	)
	if err != nil {
		return nil, err
	}

	return &AWSSession{s: s, bucket: bucket}, nil
}

// GetCredentials from the established session
func (a *AWSSession) GetCredentials() (credentials.Value, error) {
	return a.s.Config.Credentials.Get()
}

// UploadFile to a S3 bucket
func (a *AWSSession) UploadFile(localFile string, s3FilePath string) error {
	file, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return err
	}

	pa := &progressAWS{File: file, file: s3FilePath, contentLength: st.Size()}

	uploader := s3manager.NewUploader(a.s)

	_, err = uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket: aws.String(a.bucket),
		Key:    aws.String(s3FilePath),

		Body: pa,
	})

	return err
}

// GetBucket associated with a session
func (a *AWSSession) GetBucket() string {
	return a.bucket
}

type progressAWS struct {
	*os.File
	file          string
	contentLength int64
	downloaded    int64
	ticker        int
}

var _ io.Reader = (*progressAWS)(nil)
var _ io.ReaderAt = (*progressAWS)(nil)
var _ io.Seeker = (*progressAWS)(nil)

// Read file content
func (pa *progressAWS) Read(p []byte) (int, error) {
	return pa.File.Read(p)
}

// ReadAt specific offset in a file
func (pa *progressAWS) ReadAt(p []byte, off int64) (int, error) {
	n, err := pa.File.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	atomic.AddInt64(&pa.downloaded, int64(n))

	fmt.Printf("\r%v: %v%% %c", pa.file, 100*pa.downloaded/(pa.contentLength*2), pa.tick())

	return n, err
}

// Seek in a file
func (pa *progressAWS) Seek(offset int64, whence int) (int64, error) {
	return pa.File.Seek(offset, whence)
}

// Size return the file content length
func (pa *progressAWS) Size() int64 {
	return pa.contentLength
}

var ticker = `|\-/`

func (pa *progressAWS) tick() rune {
	pa.ticker = (pa.ticker + 1) % len(ticker)
	return rune(ticker[pa.ticker])
}
