package hash

import (
	"encoding/hex"
	"github.com/minio/sha256-simd"
	"github.com/zeebo/xxh3"
)

func GetBytes(domain, user, repo string) []byte {
	var bytesX = make([]byte, 0, len(domain)+len(user)+len(repo))
	bytesX = append(bytesX, []byte(domain)...)
	bytesX = append(bytesX, []byte(user)...)
	bytesX = append(bytesX, []byte(repo)...)
	return bytesX
}

func GetHash(domain, user, repo string) []byte {
	key := sha256.Sum256(GetBytes(domain, user, repo))
	return key[:]
}

func GetExtendedHash(domain, user, repo, X string) []byte {
	// 32 + length of uint64 in bytes
	var key = make([]byte, 0, 40)
	kex := GetHash(domain, user, repo)
	copy(key, kex[:])
	kexX := xxh3.Hash([]byte(X))
	key[32] = byte(0xff & kexX)
	key[33] = byte(0xff & (kexX >> 8))
	key[34] = byte(0xff & (kexX >> 16))
	key[35] = byte(0xff & (kexX >> 24))
	key[36] = byte(0xff & (kexX >> 32))
	key[37] = byte(0xff & (kexX >> 40))
	key[38] = byte(0xff & (kexX >> 48))
	key[39] = byte(0xff & (kexX >> 56))
	return key
}

func getMinioPath(domain, user, repo, version string) string {
	hash := GetHash(domain, user, repo)
	bytes := hex.EncodeToString(hash[:])
	return bytes[:4] + "/" + bytes[4:8] + "/" + bytes[8:12] + "/" + bytes[12:16] + "/" + bytes[16:] + "/" + version
}

func GetMinioGoModPath(domain, user, repo, version string) string {
	return getMinioPath(domain, user, repo, version) + ".mod"
}
func GetModuleGoSourcesPath(domain, user, repo, version string) string {
	return getMinioPath(domain, user, repo, version) + ".zip"
}

func GetMinioModuleInfoPath(domain, user, repo, version string) string {
	return getMinioPath(domain, user, repo, version) + ".info"
}

func GetMinioSumPath(domain, trail string) string {
	return hex.EncodeToString(GetHash(domain, "", trail)) + ".sum"
}

func GetLatestHash(domain, user, repo string) string {
	return hex.EncodeToString(GetHash(domain, user, repo))
}
