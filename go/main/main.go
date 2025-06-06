package main

import (
	"context"
	"fmt"
	"intermark/go/env"
	"intermark/go/flags"
	"intermark/go/router"
	"intermark/go/server"
	"intermark/go/system/git"
	"intermark/go/system/tailwind"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Data-Corruption/rlog/logger"
)

func main() {
	ctx := context.Background()

	// init logger
	level := env.Get(env.IM_LOG_LEVEL)
	if flags.PresentAny("-v", "--verbose") {
		level = "debug"
	}
	debug := level == "debug"
	log, err := logger.New("logs", level)
	if err != nil {
		fmt.Println("Error creating logger:", err)
		os.Exit(1)
	}
	defer log.Close()
	ctx = logger.IntoContext(ctx, log)
	// set up auto flush every 5 seconds if level is debug
	if debug {
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

	// get page cache size
	pc := env.Get(env.IM_PAGE_CACHE_MB)
	ipc, err := strconv.ParseInt(pc, 10, 64)
	if err != nil {
		exit("Error parsing IM_PAGE_CACHE_MB environment variable, must be an integer", err, log)
	}
	ipc = ipc * 1024 * 1024 // convert to bytes
	log.Debugf("IM_PAGE_CACHE_MB: %d", ipc)

	// get asset cache size
	ac := env.Get(env.IM_ASSET_CACHE_MB)
	iac, err := strconv.ParseInt(ac, 10, 64)
	if err != nil {
		exit("Error parsing IM_ASSET_CACHE_MB environment variable, must be an integer", err, log)
	}
	iac = iac * 1024 * 1024 // convert to bytes
	log.Debugf("IM_ASSET_CACHE_MB: %d", iac)

	edit := flags.PresentAny("-e", "--edit")

	if !edit && env.Get(env.IM_UPDATE_SECRET) == "" {
		fmt.Println("")
		fmt.Println("Warning: IM_UPDATE_SECRET environment variable is not set. This is required for automatic updates to work.")
		fmt.Println("See https://intermark.dev/p/usage/deployment for more information.")
		fmt.Println("")
		log.Warn("IM_UPDATE_SECRET environment variable is not set. This is required for automatic updates to work.")
	}

	// create router
	r, err := router.New(ctx, ipc, iac, edit, debug)
	if err != nil {
		exit("Error creating router, see logs for details", err, log)
	}

	// get address
	addr := env.Get(env.IM_ADDRESS)
	if !strings.HasPrefix(addr, ":") {
		exit("IM_ADDRESS environment variable must start with a colon (e.g. ':9292')", nil, log)
	}

	// start server
	log.Infof("Starting server on %s", addr)
	srv, err := server.New(&server.Config{
		Handler: r.Router,
		Addr:    addr,
		OnListen: func() {
			fmt.Println("Server listening on http://localhost" + addr)
			if edit {
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
