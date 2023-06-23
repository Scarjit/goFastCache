package hash

import (
	"encoding/hex"
	"github.com/minio/sha256-simd"
)

func GetBytes(stringsX []string) []byte {
	var bytesX = make([]byte, 0)
	for _, s := range stringsX {
		bytesX = append(bytesX, []byte(s)...)
	}
	return bytesX
}

func GetHash(strings []string) []byte {
	key := sha256.Sum256(GetBytes(strings))
	return key[:]
}

func getMinioPath(uri, version string) string {
	hash := GetHash([]string{uri})
	bytes := hex.EncodeToString(hash[:])
	return bytes[:4] + "/" + bytes[4:8] + "/" + bytes[8:12] + "/" + bytes[12:16] + "/" + bytes[16:] + "/" + version
}

func GetModPath(uri, version string) string {
	return getMinioPath(uri, version) + ".mod"
}
func GetZipPath(uri, version string) string {
	return getMinioPath(uri, version) + ".zip"
}

func GetInfoPath(uri, version string) string {
	return getMinioPath(uri, version) + ".info"
}

func GetSumPath(uri, trail string) string {
	return getMinioPath(uri, trail) + ".sum"
}

func GetLatestPath(uri string) string {
	return getMinioPath(uri, "latest") + ".latest"
}

func GetListPath(uri string) string {
	return getMinioPath(uri, "list") + ".list"
}
