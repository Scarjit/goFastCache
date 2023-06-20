package index

import (
	"fmt"
	"github.com/zeebo/assert"
	"goFastCache/pkg/logger"
	"testing"
	"time"
)

func Test_GetIndexSince(t *testing.T) {
	logger.InitLogger()
	now := time.Now().UTC()
	oneHourAgo := now.Add(-1 * time.Hour)
	indices, _ := getIndexSince(oneHourAgo)
	if len(indices) == 0 {
		t.Errorf("Got no indices")
	}
	fmt.Printf("Got %d indices\n", len(indices))
}

func Test_GetIndexSinceNoData(t *testing.T) {
	logger.InitLogger()
	indices, _ := getIndexSince(time.Date(3333, 06, 20, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, len(indices), 0)
}
