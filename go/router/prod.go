package router

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"intermark/go/config"
	"intermark/go/files"
	"intermark/go/html"
	"intermark/go/layout"
	"intermark/go/paths"
	"intermark/go/sins"
	"intermark/go/system/git"
	"intermark/go/system/lunrjs"

	"github.com/go-chi/chi/v5"
	"github.com/minio/sha256-simd"
)

func (r *Router) setupProdRoutes() error {
	// get cwd
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}
	// load all
	if err := r.LoadAll(cwd, ""); err != nil {
		return fmt.Errorf("error loading all: %w", err)
	}

	// serve landing page
	r.Router.Get("/", func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Encoding", "gzip")
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.Write(r.indexPage)
	})

	// serve pages
	r.Router.Get("/p/*", func(res http.ResponseWriter, req *http.Request) {
		path := filepath.Join(paths.DIST_DIR, filepath.Clean(req.URL.Path[3:]))
		if !strings.HasSuffix(path, ".html") {
			path += ".html"
		}
		data, _, zipped, err := r.pageCache.Read(path)
		if err != nil {
			r.log.Errorf("error reading page %s: %v\n", path, err)
			http.NotFound(res, req)
			return
		}
		if zipped {
			res.Header().Set("Content-Encoding", "gzip")
		}
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		res.Write(data)
	})

	// prod assets
	r.Router.Get("/a/{name}", func(res http.ResponseWriter, req *http.Request) {
		name := chi.URLParam(req, "name")
		if path, ok := r.assHashToPath[name]; !ok {
			r.log.Debugf("Asset %s not found\n", name)
			http.NotFound(res, req)
			return
		} else {
			r.log.Debugf("Serving asset %s from %s\n", name, path)
			data, mime, zipped, err := r.assetCache.Read(path[1:]) // remove leading "/"
			if err != nil {
				r.log.Errorf("Error reading asset %s: %v\n", path, err)
				http.NotFound(res, req)
				return
			}
			if zipped {
				res.Header().Set("Content-Encoding", "gzip")
			}
			res.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			res.Header().Set("Content-Type", mime)
			res.Write(data)
		}
	})

	// search index
	r.Router.Get("/search.json", func(res http.ResponseWriter, req *http.Request) {
		if match := req.Header.Get("If-None-Match"); match == r.searchHash {
			r.log.Debugf("Search index not modified, sending 304\n")
			res.Header().Set("ETag", match)
			res.WriteHeader(http.StatusNotModified)
			return
		}
		res.Header().Set("ETag", r.searchHash)
		res.Header().Set("Content-Encoding", "gzip")
		res.Header().Set("Content-Type", "application/json")
		res.Write(r.searchIdx)
	})

	// update from content repo action
	r.Router.Post("/update", func(res http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Error reading request body", http.StatusInternalServerError)
			return
		}
		// get token from env
		token := os.Getenv(UPDATE_TOKEN_KEY)
		if token == "" {
			http.Error(res, "Update token not set", http.StatusInternalServerError)
			return
		}
		// check token
		if string(body) != token {
			http.Error(res, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if err := r.Update(); err != nil {
			r.log.Errorf("Error updating: %v\n", err)
			http.Error(res, "Error updating", http.StatusInternalServerError)
			return
		}
	})

	return nil
}

func (r *Router) IsUpdating() bool {
	return r.updateFlag.Load()
}

func (r *Router) SetUpdating() error {
	if r.updateFlag.CompareAndSwap(false, true) {
		return nil
	}
	return fmt.Errorf("update already in progress")
}

func (r *Router) ClearUpdating() {
	if r.updateFlag.CompareAndSwap(true, false) {
		return
	}
	r.log.Warnf("update flag already cleared")
}

func (r *Router) Update() error {
	// set updating flag
	if err := r.SetUpdating(); err != nil {
		return fmt.Errorf("error setting updating flag: %w", err)
	}
	defer r.ClearUpdating()

	// get cwd
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %w", err)
	}

	// fetch
	fTimeout := config.GetData(r.ctx).Timeouts.Fetch
	fCtx, fCancel := context.WithTimeout(r.ctx, time.Duration(fTimeout)*time.Second)
	defer fCancel()
	if err := git.Fetch(fCtx, cwd, "main"); err != nil {
		return fmt.Errorf("error fetching latest changes: %w", err)
	}

	// reset
	rTimeout := config.GetData(r.ctx).Timeouts.Reset
	rCtx, rCancel := context.WithTimeout(r.ctx, time.Duration(rTimeout)*time.Second)
	defer rCancel()
	if err := git.Reset(rCtx, cwd, "main", true); err != nil {
		return fmt.Errorf("error resetting to latest changes: %w", err)
	}

	// lsf pull
	lTimeout := config.GetData(r.ctx).Timeouts.Lfs
	lCtx, lCancel := context.WithTimeout(r.ctx, time.Duration(lTimeout)*time.Second)
	defer lCancel()
	if err := git.LfsPull(lCtx, cwd); err != nil {
		return fmt.Errorf("error pulling LFS files: %w", err)
	}

	// get commit hash
	cCtx, cCancel := context.WithTimeout(r.ctx, 10*time.Second)
	defer cCancel()
	var newCommit string
	if newCommit, err = git.GetCommitHash(cCtx, cwd); err != nil {
		return fmt.Errorf("error getting commit hash: %w", err)
	}

	// load all
	if err := r.LoadAll(cwd, newCommit); err != nil {
		return fmt.Errorf("error loading all: %w", err)
	}

	return nil
}

// LoadAll loads all templates, assets, etc.
// If env commit / newCommit are provided and not the first LoadAll of the
// process, it will only do work for changed files.
func (r *Router) LoadAll(cwd, newCommit string) error {
	loadAllStart := time.Now()
	// get last commit hash
	distCommit := os.Getenv(DIST_CM_KEY)

	// register assets
	if err := r.registerAssets(cwd, distCommit); err != nil {
		return fmt.Errorf("error registering assets: %w", err)
	}

	// load templates, layout, and run Tailwind
	if err := r.refresh(); err != nil {
		return err
	}

	// generate dist from public
	if err := r.genDist(); err != nil {
		return fmt.Errorf("error generating dist: %w", err)
	}

	// clear cache
	r.pageCache.Reset()
	r.assetCache.Reset()

	// update commit hash
	if err := os.Setenv(DIST_CM_KEY, newCommit); err != nil {
		return fmt.Errorf("error setting environment variable %s: %w", DIST_CM_KEY, err)
	}

	// log time taken in seconds or milliseconds if < 1s
	if time.Since(loadAllStart) > time.Second {
		r.log.Debugf("LoadAll took: %s\n", time.Since(loadAllStart).String())
	} else {
		r.log.Debugf("LoadAll took: %dms\n", time.Since(loadAllStart).Milliseconds())
	}

	return nil
}

func (r *Router) registerAssets(cwd, lastCommit string) error {
	htp := sync.Map{}
	pth := sync.Map{}
	r.log.Debugf("htp len: %d", len(r.assHashToPath))
	errs, err := files.WalkCon(paths.ASS_DIR, 8, func(path string) error { // path will be `assets/example.thing`
		// asset skip check
		ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
		defer cancel()
		if changed, err := git.LFSFileChanged(ctx, cwd, path, lastCommit); err != nil {
			return err
		} else if !changed {
			// get from old, copy to new, avoid hashing
			if old_aPath, ok := r.assPathToHash[path]; ok {
				htp.Store(old_aPath, path)
				pth.Store(path, old_aPath)
				return nil
			}
		}
		// read data
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// hash and add to temp maps
		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])
		aPath := hash + filepath.Ext(path)
		path = "/" + filepath.ToSlash(filepath.Clean(path))
		htp.Store(aPath, path)
		pth.Store(path, aPath)
		r.log.Debugf("aPath: %s, path: %s\n", aPath, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error registering assets: %w", err)
	}
	if len(errs) > 0 {
		for _, we := range errs {
			r.log.Errorf("ass reg error on %s: %v\n", we.Path, we.Err)
		}
		return fmt.Errorf("%d errors registering assets, see logs for details", len(errs))
	}
	// reset the real maps
	r.assHashToPath = make(map[string]string)
	r.assPathToHash = make(map[string]string)
	// copy from the temporary sync.Maps
	htp.Range(func(key, value any) bool {
		hash := key.(string)
		path := value.(string)
		r.assHashToPath[hash] = path
		return true
	})
	pth.Range(func(key, value any) bool {
		path := key.(string)
		hash := value.(string)
		r.assPathToHash[path] = hash
		return true
	})
	return nil
}

func (r *Router) genDist() error {
	// clear dist dir
	if err := os.RemoveAll(paths.DIST_DIR); err != nil {
		return fmt.Errorf("error removing dist directory: %w", err)
	}

	// gen index
	indexPage, err := layout.Render(filepath.Join(paths.PUB_DIR, ".index.md"), r.layout.IndexTmpl, r.templates, r.layout, r.assPathToHash, r.debugMode)
	if err != nil {
		return fmt.Errorf("error processing index file: %w", err)
	}
	// compress index page
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	_, err = gz.Write([]byte(indexPage))
	if err != nil {
		return fmt.Errorf("error writing to gzip buffer: %w", err)
	}
	gz.Close()
	r.indexPage = b.Bytes()
	r.log.Debugf("Generated index page. Before gzip: %d bytes, after gzip: %d bytes\n", len(indexPage), len(r.indexPage))

	docs := []html.Doc{}

	// extract docs from index page
	if err := html.ExtractDocs("/", []byte(indexPage), &docs, sins.Ternary(r.debugMode, r.log, nil)); err != nil {
		return fmt.Errorf("error extracting docs from index page: %w", err)
	}

	// gen everything else
	errors := []error{}
	visitedItems := 0
	writeCount := 0
	r.layout.Walk(func(si *layout.SidebarItem) (bool, error) {
		visitedItems++
		if si.Type != "file" {
			return false, nil
		}
		data, err := si.Render(r.templates, r.layout, r.assPathToHash, r.debugMode)
		if err != nil {
			errors = append(errors, fmt.Errorf("error executing template: %w", err))
			return false, nil
		}
		// store in dist
		outPath := filepath.Join(paths.DIST_DIR, si.Path)
		// strip extension, add .html
		if strings.HasSuffix(outPath, ".md") {
			outPath = outPath[:len(outPath)-3] + ".html"
		}
		// ensure parent dir exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			errors = append(errors, fmt.Errorf("error creating dist directory %s: %w", outPath, err))
			return false, nil
		}
		// write to dist dir
		if err := os.WriteFile(outPath, []byte(data), 0o644); err != nil {
			errors = append(errors, fmt.Errorf("error writing file %s: %w", outPath, err))
			return false, nil
		}
		// add to search
		if err := html.ExtractDocs(si.Path, []byte(data), &docs, sins.Ternary(r.debugMode, r.log, nil)); err != nil {
			errors = append(errors, fmt.Errorf("error extracting docs from %s: %w", si.Path, err))
			return false, nil
		}
		writeCount++
		r.log.Debugf("Generated file %s, size: %d\n", outPath, len(data))
		return false, nil
	})

	// handle errors generated during walk
	if len(errors) > 0 {
		for _, err := range errors {
			r.log.Errorf("dist gen error: %v\n", err)
		}
		return fmt.Errorf("%d errors generating dist, see logs for details", len(errors))
	}

	// run lunrjs to generate search index
	lTimeout := config.GetData(r.ctx).Timeouts.Lunr
	lCtx, lCancel := context.WithTimeout(r.ctx, time.Duration(lTimeout)*time.Second)
	defer lCancel()
	if r.searchIdx, r.searchHash, err = lunrjs.Run(lCtx, &docs); err != nil {
		return fmt.Errorf("error running lunrjs: %w", err)
	}

	// log results
	r.log.Debugf("Visited %d items, wrote %d files\n", visitedItems, writeCount)
	if visitedItems == 0 {
		r.log.Warnf("No items visited, check your layout and templates")
	}
	if writeCount == 0 {
		r.log.Warnf("No files generated, check your layout and templates")
	}

	return nil
}
