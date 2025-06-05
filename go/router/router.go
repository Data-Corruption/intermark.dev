package router

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"intermark/go/captcha"
	"intermark/go/config"
	"intermark/go/files"
	"intermark/go/flags"
	"intermark/go/layout"
	"intermark/go/system/tailwind"
	"intermark/go/templates"

	"github.com/Data-Corruption/rlog/logger"
	"github.com/go-chi/chi/v5"
	"github.com/minio/sha256-simd"
)

const (
	DIST_CM_KEY      = "INTERMARK_DIST_COMMIT_KEY" // used in diff checks, set after pull/update
	UPDATE_TOKEN_KEY = "INTERMARK_UPDATE_TOKEN_KEY"
)

//go:embed updating.html
var updatingPage []byte

type Router struct {
	ctx       context.Context
	log       *logger.Logger
	Router    *chi.Mux
	templates *template.Template
	layout    *layout.Layout
	editMode  bool
	debugMode bool

	// prod stuff
	pageCache     *files.LRU
	assetCache    *files.LRU
	searchHash    string            // perm cached lunrjs index hash
	searchIdx     []byte            // perm cached lunrjs index
	indexPage     []byte            // perm cached index page
	assHashToPath map[string]string // "hash.ext" -> "/assets/example.ext"
	assPathToHash map[string]string // "/assets/example.ext" -> "hash.ext"
	updateFlag    atomic.Bool

	// edit stuff
	editMu sync.RWMutex
}

func New(ctx context.Context, pageCacheBytes, assetCacheBytes int64) (*Router, error) {
	r := &Router{
		Router:        chi.NewRouter(),
		templates:     nil,
		layout:        &layout.Layout{},
		pageCache:     files.NewLRU(true, pageCacheBytes),
		assetCache:    files.NewLRU(true, assetCacheBytes),
		assHashToPath: make(map[string]string),
		assPathToHash: make(map[string]string),
		editMode:      flags.PresentAny("-e", "--edit"),
		debugMode:     config.GetData(ctx).LogLevel == "debug",
		ctx:           ctx,
		log:           logger.FromContext(ctx),
	}

	if !r.editMode {
		// update check middleware
		r.Router.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if r.IsUpdating() {
					r.log.Debugf("Request while updating: %s\n", req.URL.Path)
					res.Header().Set("Content-Type", "text/html; charset=utf-8")
					res.WriteHeader(http.StatusServiceUnavailable)
					res.Write(updatingPage)
					return
				}
				next.ServeHTTP(res, req)
			})
		})
	}

	captcha.Setup(r.Router)

	// edit asset serving / prod fallback
	r.Router.Get("/assets/*", func(res http.ResponseWriter, req *http.Request) {
		clean := filepath.Clean(req.URL.Path)
		if !r.editMode {
			r.log.Warnf("serving static file without fingerprinting: %s\n", clean)
		}
		res.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		http.ServeFile(res, req, clean[1:]) // remove leading "/"
	})

	var err error = nil
	if r.editMode {
		r.setupEditRoutes()
	} else {
		err = r.setupProdRoutes()
	}

	return r, err
}

func (r *Router) loadTemplates() error {
	var err error
	if r.templates, err = templates.LoadTemplates(r.ctx); err != nil {
		return fmt.Errorf("error loading templates: %w", err)
	}
	return nil
}

func (r *Router) RunTailwind() error {
	tTimeout := config.GetData(r.ctx).Timeouts.Tail
	tCtx, tCancel := context.WithTimeout(r.ctx, time.Duration(tTimeout)*time.Second)
	defer tCancel()

	// run tailwind
	var out string
	var err error
	if r.editMode {
		_, err = tailwind.Run(tCtx, nil)
		return err // no need to hash in edit mode
	} else {
		out, err = tailwind.Run(tCtx, r.assPathToHash)
	}
	if err != nil {
		r.log.Debugf("error tailwind output: \n\n%s\n\n", out)
		return fmt.Errorf("error running tailwind: %w", err)
	}
	// read file and hash
	data, err := os.ReadFile(tailwind.OUTPUT_PATH)
	if err != nil {
		return fmt.Errorf("error reading tailwind output file: %w", err)
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	aPath := hash + filepath.Ext(tailwind.OUTPUT_PATH)
	r.assHashToPath[aPath] = tailwind.OUTPUT_PATH[1:] // remove leading "."
	r.assPathToHash[tailwind.OUTPUT_PATH[1:]] = aPath
	return nil
}
