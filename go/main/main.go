package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"intermark/go/config"
	"intermark/go/flags"
	"intermark/go/router"
	"intermark/go/server"
	"intermark/go/system/git"
	"intermark/go/system/tailwind"

	"github.com/Data-Corruption/rlog/logger"
)

func main() {
	ctx := context.Background()

	// init config
	cfg, err := config.New("./public/.meta/config.toml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}
	ctx = config.IntoContext(ctx, cfg)

	// init logger
	level := config.GetData(ctx).LogLevel
	if flags.PresentAny("-v", "--verbose") {
		level = "debug"
	}
	debugMode := level == "debug"
	log, err := logger.New("logs", level)
	if err != nil {
		fmt.Println("Error creating logger:", err)
		os.Exit(1)
	}
	defer log.Close()
	ctx = logger.IntoContext(ctx, log)
	// set up auto flush every 5 seconds if level is debug
	if debugMode {
		go func() {
			for {
				time.Sleep(5 * time.Second)
				if err := log.Flush(); err != nil {
					fmt.Printf("Error flushing logger: %v\nStopping auto flush", err)
					return
				}
			}
		}()
	}

	// get cwd
	cwd, err := os.Getwd()
	if err != nil {
		exit("Error getting current working directory", err, log)
	}

	// check git version
	gCtx, gCancel := context.WithTimeout(ctx, time.Second*5)
	defer gCancel()
	gVer, rVer, urVer, err := git.DebugInfo(gCtx, cwd)
	if err != nil {
		exit("Error getting git version, repo hash, or upstream hash. Is git installed?", err, log)
	}
	log.Infof("Git version: %s, repo hash: %s, upstream hash: %s", gVer, rVer, urVer)

	// check tailwind version
	ctCtx, ctCancel := context.WithTimeout(ctx, time.Second*5)
	defer ctCancel()
	tVer, err := tailwind.Version(ctCtx)
	if err != nil {
		exit("Error getting tailwind version. Try running `go run ./inter.go deps`", err, log)
	}
	log.Infof("Tailwind version: %s", tVer)

	maxPageCacheBytes := int64(config.GetData(ctx).PageCacheMB) * 1024 * 1024
	maxAssetCacheBytes := int64(config.GetData(ctx).AssetCacheMB) * 1024 * 1024
	log.Debugf("Max page cache size: %d bytes", maxPageCacheBytes)
	log.Debugf("Max asset cache size: %d bytes", maxAssetCacheBytes)

	// create router
	r, err := router.New(ctx, maxPageCacheBytes, maxAssetCacheBytes)
	if err != nil {
		exit("Error creating router, see logs for details", err, log)
	}

	// start server
	addr := config.GetData(ctx).Addr
	log.Infof("Starting server on %s", addr)
	srv, err := server.New(&server.Config{
		Handler: r.Router,
		Addr:    addr,
		OnListen: func() {
			fmt.Println("Server listening on http://localhost" + addr)
			if flags.PresentAny("-e", "--edit") {
				fmt.Println("Edit layout at http://localhost" + addr + "/edit")
			}
		},
		OnShutdown: func() {
			fmt.Println("Server shutting down...")
		},
	})
	if err != nil {
		exit("Error creating server, see logs for details", err, log)
	}
	if err = srv.Listen(); err != nil {
		exit("Error while running server, see logs for details", err, log)
	}
}

// helper that print and logs an error then exits
func exit(msg string, err error, log *logger.Logger) {
	fmt.Println(msg)
	log.Error(err)
	if err := log.Close(); err != nil {
		fmt.Println("Error closing logger:", err)
	}
	fmt.Println("See logs for details")
	fmt.Println("Exiting...")
	os.Exit(1)
}
