package selfupdate

import (
	"io"
)

type progressReader struct {
	io.Reader
	progressCallback func(float64, error)
	contentLength    int64
	downloaded       int64
}

var _ io.Reader = (*progressReader)(nil)

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.downloaded += int64(n)

	if err != io.EOF {
		if pr.contentLength > 0 {
			pr.progressCallback(float64(pr.downloaded)/float64(pr.contentLength), err)
		} else {
			pr.progressCallback(float64(pr.contentLength), err)
		}
	} else {
		pr.progressCallback(1, nil)
	}

	return n, err
}
