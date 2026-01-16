package storage

import (
	"io"

	"github.com/opencontainers/go-digest"
)

type BlobStorage interface {
	// GetUploadBlobSize return 0 if not found
	GetUploadBlobSize(id string) (int64, error)

	// UploadBlob return blob size
	UploadBlob(id string, r io.Reader) (int64, error)
	CommitBlob(repoName string, id string, d digest.Digest) error
	SaveBlob(repoName string, d digest.Digest, data []byte) error
	ReadBlob(repoName string, d digest.Digest) ([]byte, error)
	// ReadBlobToWriter write blob to w, and verify digest
	ReadBlobToWriter(repoName string, d digest.Digest, w io.Writer) (int64, error)
	IsExistBlob(repoName string, d digest.Digest) (bool, error)
	DeleteBlob(repoName string, d digest.Digest) error
}

type MetaStorage interface {
	SaveTag(repoName string, d digest.Digest, tag string) error
	ReadTag(repoName string, tag string) (string, error)
	DeleteTag(repoName string, tag string) error
	// GetTagList limit(default: -1), last: optional
	GetTagList(repoName string, limit int, last string) ([]string, error)
}
