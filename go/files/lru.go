package files

import (
	"bytes"
	"compress/gzip"
	"container/list"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LRU is a concurrent-safe fixed‐capacity read only file cache.
// It uses a least-recently-used eviction policy and is very bare bones.
type LRU struct {
	gzip  bool                     // compression enabled
	cap   int64                    // capacity in bytes
	size  int64                    // current size in bytes
	list  *list.List               // front=most recent, back=least
	cache map[string]*list.Element // path → element in ll
	mu    sync.Mutex
}

type entry struct {
	path    string
	mime    string
	data    []byte
	size    int64
	zipped  bool // true if data is compressed
	modTime time.Time
}

// NewLRU returns an [LRU] cache that will hold up to cap bytes of file data.
// If a single file’s size > cap, it will still read it but not cache it.
// Cap defaults to 1MB if <= 0.
func NewLRU(gzip bool, cap int64) *LRU {
	if cap <= 0 {
		cap = 1 << 20 // 1MB
	}
	return &LRU{
		gzip:  gzip,
		cap:   cap,
		list:  list.New(),
		cache: make(map[string]*list.Element),
	}
}

// Reset puts the cache back to its initial state.
func (l *LRU) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.list.Init()
	l.cache = make(map[string]*list.Element)
	l.size = 0
}

// Read returns the contents of the file at path, the mime, or an error. Loads from cache when possible.
// On a cache miss it reads from disk, caches it (evicting old entries if needed), and returns the bytes.
func (l *LRU) Read(path string) ([]byte, string, bool, error) {
	// check cache
	l.mu.Lock()
	if ele, ok := l.cache[path]; ok {
		l.list.MoveToFront(ele)
		ent := ele.Value.(*entry)
		size := len(ent.data)
		l.mu.Unlock()
		dataCopy := make([]byte, size)
		copy(dataCopy, ent.data)
		return dataCopy, ent.mime, ent.zipped, nil
	}
	l.mu.Unlock()

	// open file
	f, err := os.Open(path)
	if err != nil {
		return nil, "", false, err
	}
	defer f.Close()

	// get contents
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", false, err
	}

	// detect mime type
	mime := DetectMimeType(path, data)

	// compress
	zipped := false
	if l.gzip && shouldGzip(mime) {
		zipped = true
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err = gz.Write(data)
		if err != nil {
			return nil, "", false, err
		}
		gz.Close()
		data = b.Bytes()
	}

	size := int64(len(data))

	// if file is larger than capacity, don’t cache it
	if size > l.cap {
		return data, mime, zipped, nil
	}

	// get info
	info, err := f.Stat()
	if err != nil {
		return nil, "", zipped, err
	}

	// locking is separated like this to avoid lock during disk I/O
	l.mu.Lock()
	defer l.mu.Unlock()

	// re-check if file was added during disk I/O
	if ele, ok := l.cache[path]; ok {
		l.list.MoveToFront(ele)
		ent := ele.Value.(*entry)
		if ent.modTime.Before(info.ModTime()) {
			// file was modified, update cache entry
			ent.mime = mime
			copy(ent.data, data)
			ent.size = size
			ent.zipped = zipped
			ent.modTime = info.ModTime()
			l.size += size - ent.size
		} else {
			// file was not modified, return cached data
			return data, ent.mime, zipped, nil
		}
	}

	// insert new entry at front
	ent := &entry{path: path, mime: mime, data: data, size: size, zipped: zipped, modTime: info.ModTime()}
	ele := l.list.PushFront(ent)
	l.cache[path] = ele
	l.size += size

	// evict until under capacity
	for l.size > l.cap {
		back := l.list.Back()
		if back == nil {
			break
		}
		old := back.Value.(*entry)
		l.list.Remove(back)
		delete(l.cache, old.path)
		l.size -= old.size
	}

	dataCopy := make([]byte, size)
	copy(dataCopy, data)
	return dataCopy, mime, zipped, nil
}

var gzipSafeMIMEs = map[string]bool{
	"text/plain":               true,
	"text/css":                 true,
	"text/html":                true,
	"text/javascript":          true,
	"application/javascript":   true,
	"application/x-javascript": true,
	"application/json":         true,
	"application/xml":          true,
	"image/svg+xml":            true,
	"application/xhtml+xml":    true,
	"application/wasm":         true,
}

var gzipUnsafeMIMEs = map[string]bool{
	"video/mp4":          true,
	"audio/mpeg":         true,
	"audio/ogg":          true,
	"video/webm":         true,
	"image/png":          true,
	"image/jpeg":         true,
	"image/gif":          true,
	"font/woff2":         true,
	"application/zip":    true,
	"application/x-gzip": true,
}

func shouldGzip(mime string) bool {
	if gzipUnsafeMIMEs[mime] {
		return false
	}
	if safe, ok := gzipSafeMIMEs[mime]; ok {
		return safe
	}
	// fallback
	return strings.HasPrefix(mime, "text/") ||
		strings.HasSuffix(mime, "+xml") ||
		strings.HasSuffix(mime, "+json")
}
