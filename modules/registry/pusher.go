package registry

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"

	"github.com/docker/distribution"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/layer"
	"github.com/docker/docker/pkg/progress"
	"golang.org/x/net/context"
)

// v2LayerPush fulfils the docker/docker/distribution/xfer.UplaodDescriptor
// interface, allowing us to push layers to the registry.
type v2LayerPush struct {
	rand        *rand.Rand
	layerNumber int
	size        int64
	repo        distribution.Repository
	descriptor  distribution.Descriptor
}

// Key returns the key used to deduplicate uploads.
func (v v2LayerPush) Key() string {
	return fmt.Sprintf("random-layer-%d", v.layerNumber)
}

func (v v2LayerPush) ID() string {
	return fmt.Sprintf("Random layer %d", v.layerNumber)
}

func (v v2LayerPush) DiffID() layer.DiffID {
	return layer.DiffID("")
}

// Upload is called to perform the Upload.
func (v v2LayerPush) Upload(ctx context.Context, progressOutput progress.Output) (distribution.Descriptor, error) {
	bs := v.repo.Blobs(ctx)
	layerUpload, err := bs.Create(ctx)
	if err != nil {
		return distribution.Descriptor{}, err
	}

	// In docker/docker we recreate the .tar.gz based on tar assembly from
	// metadata and the layer information. Here, we'll create a new tar.gz
	// with base layer data that has no info.
	tarReader, err := createTar(v.rand, v.size)
	if err != nil {
		return distribution.Descriptor{}, err
	}

	// Compress the tar and create a hash at the same time
	compressReader, compressDone := compress(tarReader)
	digester := digest.Canonical.New()
	tee := io.TeeReader(compressReader, digester.Hash())

	go func() {
		<-compressDone
	}()

	nn, err := layerUpload.ReadFrom(tee)
	compressReader.Close()
	if err != nil {
		return distribution.Descriptor{}, err
	}

	pushDigest := digester.Digest()
	if _, err := layerUpload.Commit(ctx, distribution.Descriptor{Digest: pushDigest}); err != nil {
		return distribution.Descriptor{}, err
	}

	return distribution.Descriptor{
		Digest:    pushDigest,
		MediaType: schema2.MediaTypeLayer,
		Size:      nn,
	}, nil
}

func createTar(r *rand.Rand, uncompressedSize int64) (io.Reader, error) {
	// Create a new TAR
	tBuf := new(bytes.Buffer)
	tarWriter := tar.NewWriter(tBuf)
	hdr := &tar.Header{
		Name: "/rand",
		Mode: 0400,
		Size: uncompressedSize,
	}
	err := tarWriter.WriteHeader(hdr)
	if err != nil {
		return nil, err
	}

	// Only read N bytes from rand
	limitedRand := io.LimitReader(r, uncompressedSize)
	// Read from limitedRand and write to the .tar file immediately
	_, err = ioutil.ReadAll(io.TeeReader(limitedRand, tarWriter))
	if err != nil {
		return nil, err
	}

	err = tarWriter.Close()
	if err != nil {
		return nil, err
	}

	return tBuf, nil
}

// SetRemoteDescriptor provides the distribution.Descriptor that was
// returned by Upload. This descriptor is not to be confused with
// the UploadDescriptor interface, which is used for internally
// identifying layers that are being uploaded.
func (v *v2LayerPush) SetRemoteDescriptor(descriptor distribution.Descriptor) {
	fmt.Println("SETTING DESCRIPTOR", descriptor)
	v.descriptor = descriptor
}

func (v *v2LayerPush) Descriptor() distribution.Descriptor {
	return v.descriptor
}

const compressionBufSize = 2 ^ 15

// compress returns an io.ReadCloser which will supply a compressed version of
// the provided Reader. The caller must close the ReadCloser after reading the
// compressed data.
//
// Note that this function returns a reader instead of taking a writer as an
// argument so that it can be used with httpBlobWriter's ReadFrom method.
// Using httpBlobWriter's Write method would send a PATCH request for every
// Write call.
//
// The second return value is a channel that gets closed when the goroutine
// is finished. This allows the caller to make sure the goroutine finishes
// before it releases any resources connected with the reader that was
// passed in.
func compress(in io.Reader) (io.ReadCloser, chan struct{}) {
	compressionDone := make(chan struct{})

	pipeReader, pipeWriter := io.Pipe()
	// Use a bufio.Writer to avoid excessive chunking in HTTP request.
	bufWriter := bufio.NewWriterSize(pipeWriter, compressionBufSize)
	compressor := gzip.NewWriter(bufWriter)

	go func() {
		_, err := io.Copy(compressor, in)
		if err == nil {
			err = compressor.Close()
		}
		if err == nil {
			err = bufWriter.Flush()
		}
		if err != nil {
			pipeWriter.CloseWithError(err)
		} else {
			pipeWriter.Close()
		}
		close(compressionDone)
	}()

	return pipeReader, compressionDone
}
