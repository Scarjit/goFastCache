package hash

import (
	"crypto/sha256"
	"fmt"
	"github.com/cespare/xxhash"
	simd256 "github.com/minio/sha256-simd"
	"github.com/zeebo/xxh3"
	"golang.org/x/crypto/sha3"
	"testing"
)

func GetHashSHA256(uri string) [32]byte {
	return sha256.Sum256(GetBytes(uri))
}

func GetHashSHA256SIMD(uri string) [32]byte {
	return simd256.Sum256(GetBytes(uri))
}

func GetHashSHA3_256(uri string) []byte {
	return sha3.New256().Sum(GetBytes(uri))
}

func GetHashXXHash64(uri string) uint64 {
	return xxhash.Sum64(GetBytes(uri))
}

func GetHashXXHash3(uri string) uint64 {
	return xxh3.Hash(GetBytes(uri))
}

func GetHashXXHash3_128(uri string) [16]byte {
	return xxh3.Hash128(GetBytes(uri)).Bytes()
}

const uri = "v1.0.0"
const user = "user"
const repo = "repo"

func BenchmarkSHA256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA256(uri)
	}
}

func BenchmarkSHA256SIMD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA256SIMD(uri)
	}
}

func BenchmarkSHA3_256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA3_256(uri)
	}
}

func BenchmarkXXHash64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash64(uri)
	}
}

func BenchmarkXXHash3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash3(uri)
	}
}

func BenchmarkXXHash3_128(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash3_128(uri)
	}
}

func TestGetMinioPath(t *testing.T) {
	path := getMinioPath(uri, "v1.0.0")
	fmt.Printf("path: %s\n", path)
}
