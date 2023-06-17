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

func GetHashSHA256(domain, user, repo string) [32]byte {
	return sha256.Sum256(GetBytes(domain, user, repo))
}

func GetHashSHA256SIMD(domain, user, repo string) [32]byte {
	return simd256.Sum256(GetBytes(domain, user, repo))
}

func GetHashSHA3_256(domain, user, repo string) []byte {
	return sha3.New256().Sum(GetBytes(domain, user, repo))
}

func GetHashXXHash64(domain, user, repo string) uint64 {
	return xxhash.Sum64(GetBytes(domain, user, repo))
}

func GetHashXXHash3(domain, user, repo string) uint64 {
	return xxh3.Hash(GetBytes(domain, user, repo))
}

func GetHashXXHash3_128(domain, user, repo string) [16]byte {
	return xxh3.Hash128(GetBytes(domain, user, repo)).Bytes()
}

const domain = "v1.0.0"
const user = "user"
const repo = "repo"

func BenchmarkSHA256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA256(domain, user, repo)
	}
}

func BenchmarkSHA256SIMD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA256SIMD(domain, user, repo)
	}
}

func BenchmarkSHA3_256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashSHA3_256(domain, user, repo)
	}
}

func BenchmarkXXHash64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash64(domain, user, repo)
	}
}

func BenchmarkXXHash3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash3(domain, user, repo)
	}
}

func BenchmarkXXHash3_128(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetHashXXHash3_128(domain, user, repo)
	}
}

func TestGetMinioPath(t *testing.T) {
	path := GetMinioPath(domain, user, repo)
	fmt.Printf("path: %s\n", path)
}
