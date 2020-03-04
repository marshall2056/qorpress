package helpers

import (
	"os"
	"syscall"
	"time"
)

func statTimes(fi os.FileInfo) (atime, mtime, ctime time.Time, err error) {
	mtime = fi.ModTime()
	stat := fi.Sys().(*syscall.Stat_t)
	atime = time.Unix(int64(stat.Atimespec.Sec), int64(stat.Atimespec.Nsec))
	ctime = time.Unix(int64(stat.Ctimespec.Sec), int64(stat.Ctimespec.Nsec))
	return
}
