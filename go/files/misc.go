package files

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type WalkError struct {
	Path string
	Err  error
}

// WalkCon walks the directory rooted at rootPath,
// calling work(path) concurrently for each file.
// Returns a slice of errors from work calls.
func WalkCon(rootPath string, workerCount int, work func(string) error) ([]WalkError, error) {
	paths := make(chan string, workerCount*2)
	errs := make(chan WalkError, workerCount*10)

	var wg sync.WaitGroup

	// start worker goroutines
	for i := 0; i < workerCount; i++ {
		go func() {
			for path := range paths {
				if err := work(path); err != nil {
					errs <- WalkError{Path: path, Err: err}
				}
				wg.Done()
			}
		}()
	}

	// walk the directory and send paths to workers
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			wg.Add(1)
			paths <- path
		}
		return nil
	})
	if err != nil {
		close(paths)
		wg.Wait()
		close(errs)
		return nil, err
	}

	wg.Wait()
	close(paths)
	close(errs)

	var collected []WalkError
	for fe := range errs {
		collected = append(collected, fe)
	}
	return collected, nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// FirstExists returns the first path that exists from the provided list.
// It returns the path and true if any Stat succeeds, or "" and false otherwise.
func FirstExists(dir string, paths ...string) (string, bool) {
	for _, p := range paths {
		if _, err := os.Stat(filepath.Join(dir, p)); err == nil {
			return p, true
		}
	}
	return "", false
}

func DetectMimeType(path string, data []byte) string {
	if ext := filepath.Ext(path); ext != "" {
		if byExt := mime.TypeByExtension(ext); byExt != "" {
			return byExt
		}
	}
	return http.DetectContentType(data) // fallback
}
