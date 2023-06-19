package blobstorage

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/goccy/go-json"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/sha256-simd"
	"go.uber.org/zap"
	"goFastCache/pkg/hash"
	"os"
	"strings"
)

type Blobstore struct {
	MinioClient *minio.Client
	BucketName  string
}

func NewBlobstore() (*Blobstore, error) {
	zap.S().Info("Connecting to Minio")
	minioAccessKey, found := os.LookupEnv("MINIO_ACCESS_KEY")
	if !found {
		return nil, errors.New("MINIO_ACCESS_KEY not found")
	}
	minioAccessKey = strings.Trim(minioAccessKey, "\n\r")

	minioSecretKey, found := os.LookupEnv("MINIO_SECRET_KEY")
	if !found {
		return nil, errors.New("MINIO_SECRET_KEY not found")
	}
	minioSecretKey = strings.Trim(minioSecretKey, "\n\r")

	minioBucketName, found := os.LookupEnv("MINIO_BUCKET_NAME")
	if !found {
		return nil, errors.New("MINIO_BUCKET_NAME not found")
	}
	minioBucketName = strings.Trim(minioBucketName, "\n\r")

	minioDomain, found := os.LookupEnv("MINIO_DOMAIN")
	if !found {
		return nil, errors.New("MINIO_DOMAIN not found")
	}
	minioDomain = strings.Trim(minioDomain, "\n\r")

	minioClient, err := minio.New(minioDomain, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	b := &Blobstore{
		BucketName:  minioBucketName,
		MinioClient: minioClient,
	}
	err = b.createBucket()
	if err != nil {
		return nil, err
	}
	zap.S().Info("Connected to Minio")
	return b, nil
}

func (b *Blobstore) createBucket() error {
	exists, err := b.MinioClient.BucketExists(context.Background(), b.BucketName)
	if err != nil {
		return err
	}
	if exists {
		zap.S().Info("Bucket already exists")
		return nil
	}
	err = b.MinioClient.MakeBucket(context.Background(), b.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		return err
	}
	zap.S().Infof("Bucket %s created successfully", b.BucketName)
	return nil
}

func (b *Blobstore) putString(object string, path string) (info minio.UploadInfo, err error) {
	return b.putStream([]byte(object), path, "text/plain")
}

func (b *Blobstore) putBytes(object []byte, path string) (info minio.UploadInfo, err error) {
	return b.putStream(object, path, "application/octet-stream")
}

func (b *Blobstore) putJSON(object interface{}, path string) (info minio.UploadInfo, err error) {
	marshalled, err := json.Marshal(object)
	if err != nil {
		return minio.UploadInfo{}, err
	}
	return b.putStream(marshalled, path, "application/json")
}

func (b *Blobstore) putStream(object []byte, path, contentType string) (info minio.UploadInfo, err error) {
	sha256Sum := sha256.Sum256(object)
	options := minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"X-Amz-Meta-Sha256": hex.EncodeToString(sha256Sum[:]),
		},
		ContentType: contentType,
	}

	return b.MinioClient.PutObject(context.Background(), b.BucketName, path, bytes.NewReader(object), int64(len(object)), options)
}

func (b *Blobstore) getObject(path string) (object []byte, err error) {
	var objectX *minio.Object
	objectX, err = b.MinioClient.GetObject(context.Background(), b.BucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	objectInfo, err := objectX.Stat()
	if err != nil {
		return nil, err
	}
	object = make([]byte, objectInfo.Size)
	var read int
	read, err = objectX.Read(object)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	if int64(read) != objectInfo.Size {
		return nil, errors.New("unable to read entire object")
	}

	// Verify the checksum if it exists
	sha256Sum := objectInfo.Metadata.Get("X-Amz-Meta-Sha256")
	if sha256Sum == "" {
		return nil, errors.New("checksum not found")
	}

	// Calculate the checksum
	sha256SumCalculated := sha256.Sum256(object)
	sha256SumCalculatedString := hex.EncodeToString(sha256SumCalculated[:])

	// Compare the checksums
	if sha256Sum != sha256SumCalculatedString {
		return nil, errors.New("checksums do not match")
	}

	return object, nil
}

func (b *Blobstore) removeObject(path string) error {
	return b.MinioClient.RemoveObject(context.Background(), b.BucketName, path, minio.RemoveObjectOptions{})
}

func (b *Blobstore) GetInfoObject(domain, user, repo, version string) (object map[string]interface{}, found bool, err error) {
	path := hash.GetMinioModuleInfoPath(domain, user, repo, version)
	obj, err := b.getObject(path)
	// Check if err is not found
	if err == nil {
		// Unmarshal the object
		err = json.Unmarshal(obj, &object)
		return object, true, err
	}
	if strings.Contains(err.Error(), "The specified key does not exist") {
		return nil, false, nil
	}
	return nil, false, err
}

func (b *Blobstore) PutInfoObject(domain, user, repo, version string, object []byte) error {
	path := hash.GetMinioModuleInfoPath(domain, user, repo, version)
	var err error
	_, err = b.putBytes(object, path)
	return err

}

func (b *Blobstore) GetModObject(domain, user, repo, version string) (object []byte, found bool, err error) {
	path := hash.GetMinioGoModPath(domain, user, repo, version)
	obj, err := b.getObject(path)

	if err == nil {
		return obj, true, nil
	}

	if strings.Contains(err.Error(), "The specified key does not exist") {
		return nil, false, nil
	}

	return nil, false, err
}

func (b *Blobstore) PutModObject(domain, user, repo, version string, object []byte) error {
	path := hash.GetMinioGoModPath(domain, user, repo, version)
	var err error
	_, err = b.putBytes(object, path)
	return err
}

func (b *Blobstore) GetModuleSourceObject(domain, user, repo, version string) (object []byte, found bool, err error) {
	path := hash.GetModuleGoSourcesPath(domain, user, repo, version)
	obj, err := b.getObject(path)

	if err == nil {
		return obj, true, nil
	}

	if strings.Contains(err.Error(), "The specified key does not exist") {
		return nil, false, nil
	}

	return nil, false, err
}

func (b *Blobstore) PutModuleSourceObject(domain, user, repo, version string, object []byte) error {
	path := hash.GetModuleGoSourcesPath(domain, user, repo, version)
	var err error
	_, err = b.putBytes(object, path)
	return err
}

func (b *Blobstore) GetSumObject(domain, trail string) (object []byte, found bool, err error) {
	path := hash.GetMinioSumPath(domain, trail)
	obj, err := b.getObject(path)

	if err == nil {
		return obj, true, nil
	}

	if strings.Contains(err.Error(), "The specified key does not exist") {
		return nil, false, nil
	}

	return nil, false, err
}

func (b *Blobstore) PutSumObject(domain, trail string, object []byte) error {
	path := hash.GetMinioSumPath(domain, trail)
	var err error
	_, err = b.putBytes(object, path)
	return err
}
