package storage

import (
	"io"

	"github.com/opencontainers/go-digest"
)

type Storage interface {
	// return 0 if not found
	GetUploadBlobSize(id string) (int64, error)

	// return blob size
	UploadBlob(id string, r io.Reader) (int64, error)
	CommitBlob(repoName string, id string, d digest.Digest) error
	SaveBlob(repoName string, d digest.Digest, data []byte) error
	ReadBlob(repoName string, d digest.Digest) ([]byte, error)
	ReadBlobToWriter(repoName string, d digest.Digest, w io.Writer) (int64, error)
	IsExistBlob(repoName string, d digest.Digest) (bool, error)

	SaveTag(repoName string, d digest.Digest, tag string) error
	ReadTag(repoName string, tag string) (string, error)
	GetTagList(repoName string) ([]string, error)
}
