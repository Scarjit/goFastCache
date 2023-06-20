package index

import (
	"bufio"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"time"
)

func GetIndexSince(since time.Time) []Index {
	// https://index.golang.org/index?since=2023-06-20T00:00:00.000000Z
	// We can only download 2000 packages at a time, so we need to do this in a loop, changing the since param

	var nextSince *time.Time
	nextSince = &since
	var err error
	var indices []Index
	for {
		var indicesTemp []Index
		nextSince, indices, err = downloadIndex(&since)
		indicesTemp = append(indicesTemp, indices...)
		if nextSince != nil || err != nil || *nextSince == since {
			break
		}
		since = *nextSince
	}
	if err != nil {
		zap.S().Errorf("Error downloading index: %v", err)
	}

	return DedupIndex(indices)
}

func DedupIndex(indices []Index) []Index {
	indexMap := make(map[string]Index)
	for _, index := range indices {
		v, ok := indexMap[index.Path]
		if !ok {
			indexMap[index.Path] = index
			continue
		}
		x, err := semver.NewVersion(index.Version)
		if err != nil {
			zap.S().Errorf("Error parsing version: %v", err)
			continue
		}
		y, err := semver.NewVersion(v.Version)
		if err != nil {
			zap.S().Errorf("Error parsing version: %v", err)
			continue
		}
		if x.GreaterThan(y) {
			indexMap[index.Path] = index
		}
	}
	indices = make([]Index, 0, len(indexMap))

	for _, index := range indexMap {
		indices = append(indices, index)
	}
	return indices
}

type Index struct {
	Path      string
	Version   string
	Timestamp time.Time
}

func downloadIndex(since *time.Time) (nextSince *time.Time, indices []Index, err error) {
	// convert since to string YYYY-MM-DDTHH:MM:SS.MSZ

	timeString := since.Format("2006-01-02T15:04:05.000000Z")
	url := fmt.Sprintf("https://index.golang.org/index?since=%s", timeString)
	zap.S().Infof("Downloading index since %s", timeString)

	// download index
	get, err := http.Get(url)
	if err != nil {
		return nil, indices, err
	}
	// Check status code
	if get.StatusCode != 200 {
		return nil, indices, fmt.Errorf("status code %d", get.StatusCode)
	}

	indices = make([]Index, 0, 2000)

	// Read as buffer, to parse JSONL
	defer get.Body.Close()
	s := bufio.NewScanner(get.Body)
	for s.Scan() {
		var index Index
		err = json.Unmarshal(s.Bytes(), &index)
		if err != nil {
			zap.S().Errorf("Error parsing JSONL: %v", err)
			err = nil
			continue
		}
		indices = append(indices, index)
	}
	// Sort by timestamp

	sort.Slice(indices, func(i, j int) bool {
		return indices[i].Timestamp.Before(indices[j].Timestamp)
	})

	// Get last timestamp
	if len(indices) > 0 {
		last := indices[len(indices)-1]
		nextSince = &last.Timestamp
	}

	return nextSince, indices, nil
}
