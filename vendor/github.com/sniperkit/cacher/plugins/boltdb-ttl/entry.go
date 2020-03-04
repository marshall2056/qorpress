package boltdbttlcache

import (
	"math"
	"time"
	// log "github.com/sirupsen/logrus"
	// "github.com/sniperkit/httpcache/helpers"
)

type IndexEntry struct {
	Key       []byte
	ExpiresAt time.Time
}

type ByIndexEntryExpiry []IndexEntry

func (s ByIndexEntryExpiry) Len() int {
	return len(s)
}
func (s ByIndexEntryExpiry) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByIndexEntryExpiry) Less(i, j int) bool {
	return s[i].ExpiresAt.Before(s[j].ExpiresAt)
}

type CacheEntry struct {
	Key       []byte
	Value     interface{}
	CreatedAt time.Time
	ExpiresAt time.Time
}

func appendIndex(index []IndexEntry, entry IndexEntry) []IndexEntry {
	newIdx := insertSortedIndexEntry(index, entry, 0, 0)
	return newIdx
}

func insertSortedIndexEntry(index []IndexEntry, entry IndexEntry, start int, end int) []IndexEntry {
	if index == nil {
		return nil
	}
	length := len(index)
	if end == 0 {
		end = length - 1
	}
	mid := start + int(math.Floor(float64(end-start)/2))
	if length == 0 {
		return append(index, entry)
	}
	firstEntry := index[end]
	lastEntry := index[end]
	middleEntry := index[mid]
	if entry.ExpiresAt.Before(firstEntry.ExpiresAt) {
		return append([]IndexEntry{entry}, index...)
	}
	if entry.ExpiresAt.After(lastEntry.ExpiresAt) || entry.ExpiresAt.Equal(lastEntry.ExpiresAt) {
		return append(index, entry)
	}
	if entry.ExpiresAt.Before(middleEntry.ExpiresAt) {
		return insertSortedIndexEntry(index, entry, start, mid-1)
	}
	if entry.ExpiresAt.After(middleEntry.ExpiresAt) || entry.ExpiresAt.Equal(middleEntry.ExpiresAt) {
		return insertSortedIndexEntry(index, entry, mid+1, end)
	}
	return index
}
