package cloudstorage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"go4.org/readerutil"
)

type GCSReaderAt struct {
	ctx    context.Context
	client *storage.Client
	bucket string
	object string
	attrs  *storage.ObjectAttrs
}

func NewGCSReaderAt(ctx context.Context, client *storage.Client, bucket, object string) (*GCSReaderAt, error) {
	attrs, err := client.Bucket(bucket).Object(object).Attrs(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSReaderAt{
		ctx:    ctx,
		client: client,
		bucket: bucket,
		object: object,
		attrs:  attrs,
	}, nil
}

func (r *GCSReaderAt) ReadAt(p []byte, off int64) (int, error) {
	rc, err := r.client.Bucket(r.bucket).Object(r.object).NewRangeReader(r.ctx, off, int64(len(p)))
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	return io.ReadFull(rc, p)
}

func (r *GCSReaderAt) Size() int64 {
	return r.attrs.Size
}

var _ readerutil.SizeReaderAt = (*GCSReaderAt)(nil)
