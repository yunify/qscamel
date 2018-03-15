package aliyun

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/sirupsen/logrus"

	"github.com/yunify/qscamel/model"
)

// Reachable implement source.Reachable
func (c *Client) Reachable() bool {
	return true
}

// Readable implement source.Readable
func (c *Client) Readable() bool {
	return true
}

// List implement source.List
func (c *Client) List(ctx context.Context, j *model.Job, rc chan *model.Object) {
	defer close(rc)

	// Add "/" to list specific prefix.
	cp := path.Join(c.Path, j.Path) + "/"
	// Trim left "/" to prevent object start with "/"
	cp = strings.TrimLeft(cp, "/")

	marker := j.Marker

	for {
		resp, err := c.client.ListObjects(
			oss.Delimiter("/"),
			oss.Marker(marker),
			oss.MaxKeys(MaxKeys),
			oss.Prefix(cp),
		)
		if err != nil {
			logrus.Errorf("List objects failed for %v.", err)
			rc <- nil
			return
		}
		for _, v := range resp.Objects {
			object := &model.Object{
				Key:   strings.TrimLeft(v.Key, c.Path),
				IsDir: false,
				Size:  v.Size,
			}

			rc <- object
		}
		for _, v := range resp.CommonPrefixes {
			object := &model.Object{
				Key:   strings.TrimLeft(v, c.Path),
				IsDir: true,
				Size:  0,
			}

			rc <- object
		}

		marker = resp.NextMarker

		// Update task content.
		j.Marker = marker
		err = j.Save(ctx)
		if err != nil {
			logrus.Errorf("Save task failed for %v.", err)
			rc <- nil
			return
		}

		if marker == "" {
			break
		}
	}

	return
}

// Read implement source.Read
func (c *Client) Read(ctx context.Context, p string) (r io.ReadCloser, err error) {
	cp := path.Join(c.Path, p)
	// Trim left "/" to prevent object start with "/"
	cp = strings.TrimLeft(cp, "/")

	r, err = c.client.GetObject(cp)
	if err != nil {
		return
	}

	return
}

// Reach implement source.Fetch
func (c *Client) Reach(ctx context.Context, p string) (url string, err error) {
	return
}
