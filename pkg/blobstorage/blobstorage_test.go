package blobstorage

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/sha256-simd"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func envFileToEnv() error {
	// Invoke git to get repo root
	// git rev-parse --show-toplevel

	command := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := command.Output()
	if err != nil {
		fmt.Printf("Error running git rev-parse: %v\n", err)
		return err
	}
	// Trim newline from output
	repoRoot := strings.TrimSuffix(string(output), "\n")
	// Set env file path
	envFile := filepath.Join(repoRoot, ".env.tests")

	envFileString, err := os.ReadFile(envFile)
	if err != nil {
		fmt.Printf("Error reading .env file: %v\n", err)
		return err
	}
	// Parse line by line, split on =, and set env vars
	for _, line := range strings.Split(string(envFileString), "\n") {
		if strings.Contains(line, "=") {
			splitLine := strings.Split(line, "=")
			err = os.Setenv(splitLine[0], splitLine[1])
			if err != nil {
				fmt.Printf("Error setting env var %s: %v\n", splitLine[0], err)
				return err
			}

			// Test if env var was set
			env, b := os.LookupEnv(splitLine[0])
			if !b || env != splitLine[1] {
				fmt.Printf("Error setting env var %s: %v\n", splitLine[0], err)
				return err
			}
			fmt.Printf("Set env var %s to %s\n", splitLine[0], env)
		}
	}

	return nil
}

func TestBlobstore_PutObject(t *testing.T) {
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	blobstore, err := NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
	testData := "test data"
	var info minio.UploadInfo

	// Test PutString
	info, err = blobstore.PutString(testData, "testobject/a/b/c/d/PutString.txt")
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	fmt.Printf("UploadInfo: %+v\n", info)

	// Test PutBytes
	info, err = blobstore.PutBytes([]byte(testData), "testobject/a/b/c/d/PutBytes.txt")
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	fmt.Printf("UploadInfo: %+v\n", info)

	// Test PutJSON
	type TestStruct struct {
		TestString string
		TestInt    int
	}
	testStruct := TestStruct{
		TestString: "test string",
		TestInt:    123,
	}
	info, err = blobstore.PutJSON(testStruct, "testobject/a/b/c/d/PutJSON.json")
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	fmt.Printf("UploadInfo: %+v\n", info)

	fmt.Println("Successfully put object")
}

func TestBlobstore_NewBlobstore(t *testing.T) {
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	_, err = NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
}

func TestBlobstore_GetObject(t *testing.T) {
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	blobstore, err := NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
	testData := "test data for get object"
	path := "testgetobject/a/b/c/d/GetObject.txt"
	_, err = blobstore.PutString(testData, path)
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	_, err = blobstore.GetObject(path)
	if err != nil {
		t.Fatalf("Error getting object: %v", err)
	}
}

func TestBlobstore_RemoveObject(t *testing.T) {
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	blobstore, err := NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
	testData := "test data for remove object"
	path := "testremoveobject/a/b/c/d/RemoveObject.txt"
	_, err = blobstore.PutString(testData, path)
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	err = blobstore.RemoveObject(path)
	if err != nil {
		t.Fatalf("Error removing object: %v", err)
	}
	_, err = blobstore.GetObject(path)
	if err == nil {
		t.Fatalf("Object was not removed")
	}
}

func TestBlobstore_NoChecksum(t *testing.T) {
	// Basically PutObject, and then modify the MinioClient directly without setting the checksum
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	blobstore, err := NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
	testData := "test data"
	var info minio.UploadInfo

	// Test PutString
	info, err = blobstore.PutString(testData, "testshaobject/a/b/c/d/PutStringWrongHash.txt")
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	fmt.Printf("UploadInfo: %+v\n", info)

	testData = "test data 2"
	// Modify the MinioClient directly to return a different checksum
	info, err = blobstore.MinioClient.PutObject(context.Background(), blobstore.BucketName, "testshaobject/a/b/c/d/PutStringWrongHash.txt", strings.NewReader(testData), int64(len(testData)), minio.PutObjectOptions{})
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}

	// Get the object and check the checksum
	_, err = blobstore.GetObject("testshaobject/a/b/c/d/PutStringWrongHash.txt")
	if err == nil || err.Error() != "checksum not found" {
		t.Fatalf("Expected 'checksum not found' error")
	}
}

func TestBlobstore_WrongChecksum(t *testing.T) {
	// Basically PutObject, and then modify the MinioClient directly to return a different checksum
	err := envFileToEnv()
	if err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
	blobstore, err := NewBlobstore()
	if err != nil {
		t.Fatalf("Error creating blobstore: %v", err)
	}
	testData := "test data"
	var info minio.UploadInfo

	// Test PutString
	info, err = blobstore.PutString(testData, "testshaobject/a/b/c/d/PutStringWrongHash.txt")
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}
	fmt.Printf("UploadInfo: %+v\n", info)

	testData2 := "test data 2"
	// Modify the MinioClient directly to return a different checksum
	sha256Sum := sha256.Sum256([]byte(testData))
	options := minio.PutObjectOptions{
		UserMetadata: map[string]string{
			"X-Amz-Meta-Sha256": hex.EncodeToString(sha256Sum[:]),
		},
		ContentType: "plain/text",
	}
	info, err = blobstore.MinioClient.PutObject(context.Background(), blobstore.BucketName, "testshaobject/a/b/c/d/PutStringWrongHash.txt", strings.NewReader(testData2), int64(len(testData2)), options)
	if err != nil {
		t.Fatalf("Error putting object: %v", err)
	}

	// Get the object and check the checksum
	_, err = blobstore.GetObject("testshaobject/a/b/c/d/PutStringWrongHash.txt")
	if err == nil || err.Error() != "checksums do not match" {
		t.Fatalf("Expected SHA256 mismatch error")
	}
	fmt.Printf("Error: %v\n", err)
}
