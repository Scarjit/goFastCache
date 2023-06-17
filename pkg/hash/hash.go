package hash

import (
	"encoding/hex"
	"github.com/minio/sha256-simd"
)

func GetBytes(domain, user, repo string) []byte {
	var bytesX = make([]byte, 0, len(domain)+len(user)+len(repo))
	bytesX = append(bytesX, []byte(domain)...)
	bytesX = append(bytesX, []byte(user)...)
	bytesX = append(bytesX, []byte(repo)...)
	return bytesX
}

func GetHash(domain, user, repo string) [32]byte {
	return sha256.Sum256(GetBytes(domain, user, repo))
}

func GetMinioPath(domain, user, repo, version string) string {
	hash := GetHash(domain, user, repo)
	bytes := hex.EncodeToString(hash[:])
	return bytes[:4] + "/" + bytes[4:8] + "/" + bytes[8:12] + "/" + bytes[12:16] + "/" + bytes[16:] + "/" + version
}
