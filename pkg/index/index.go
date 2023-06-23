package index

import (
	"bufio"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"goFastCache/pkg/blobstorage"
	"goFastCache/pkg/database"
	"goFastCache/pkg/routes"
	"net/http"
	"regexp"
	"sort"
	"time"
)

func getIndexSince(since time.Time) ([]Index, *time.Time) {
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

	return dedupIndex(indices), nextSince
}

func dedupIndex(indices []Index) []Index {
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

func RefreshIndexInBackground(db *database.Database, blob *blobstorage.Blobstore) {
	for i := 0; i < 10; i++ {
		go worker(blob)
	}
	go func() {
		now := time.Now().Add(-1 * time.Hour)
		for {
			RefreshIndex(db, now)
			time.Sleep(1 * time.Hour)
		}
	}()
}

var PathRegex = regexp.MustCompile(`(.+?/)(.+)`)

type workload struct {
	Domain     string
	ModuleName string
	Version    string
}

var workerChan = make(chan workload, 10)

func RefreshIndex(db *database.Database, refreshStart time.Time) {
	var indices []Index
	var nextStart *time.Time
	indices, nextStart = getIndexSince(refreshStart)
	if nextStart == nil {
		nextStart = &refreshStart
	}

	for _, index := range indices {
		// Check if m exists in database by path
		// If it does, check if version is newer
		m, found, err := db.GetGoModuleByPath(index.Path)
		if err != nil {
			zap.S().Errorf("Error getting m by path: %v", err)
			continue
		}
		if !found {
			continue
		}
		// Check if version is newer
		x, err := semver.NewVersion(index.Version)
		if err != nil {
			zap.S().Errorf("Error parsing version: %v", err)
			continue
		}
		y, err := semver.NewVersion(m.Version)
		if err != nil {
			zap.S().Errorf("Error parsing version: %v", err)
			continue
		}
		if x.GreaterThan(y) {
			// Update version
			m.Version = index.Version
			err = db.UpsertGoModule(m)
			if err != nil {
				zap.S().Errorf("Error updating m: %v", err)
				continue
			}
			// Trigger update of m into blob storage
			matches := PathRegex.FindStringSubmatch(m.Path)
			if len(matches) != 3 {
				zap.S().Errorf("Error parsing m path: %s", m.Path)
				continue
			}
			uri := matches[1]
			moduleName := matches[2]
			workerChan <- workload{
				Domain:     uri,
				ModuleName: moduleName,
				Version:    m.Version,
			}
		}
	}
}

func worker(blob *blobstorage.Blobstore) {
	for {
		w := <-workerChan
		routes.
	}
}
