package server

import (
	"net/http"
	"os"
	"path/filepath"
)

type ServerHandler struct {
	*Config
	mainFileExternalServer string // eg: main.server.go
	strategy               ServerStrategy
	inMemory               bool // true if running in-memory
	// goCompiler and goRun are now managed by externalStrategy
}

type Config struct {
	AppRootDir                  string                 // e.g., /home/user/project (application root directory)
	SourceDir                   string                 // directory location of main.go e.g., src/cmd/appserver (relative to AppRootDir)
	OutputDir                   string                 // compilation and execution directory e.g., deploy/appserver (relative to AppRootDir)
	PublicDir                   string                 // default public dir for generated server (e.g., src/web/public)
	MainInputFile               string                 // main input file name (default: "main.go", can be "server.go", etc.)
	ArgumentsForCompilingServer func() []string        // e.g., []string{"-X 'main.version=v1.0.0'"}
	ArgumentsToRunServer        func() []string        // e.g., []string{"dev"}
	AppPort                     string                 // e.g., 8080
	Routes                      []func(*http.ServeMux) // Functions to register routes on the HTTP server
	Logger                      func(message ...any)   // For logging output
	ExitChan                    chan bool              // Global channel to signal shutdown
}

// NewConfig provides a default configuration.
func NewConfig() *Config {
	return &Config{
		AppRootDir:    ".",
		SourceDir:     "web",
		OutputDir:     "web",
		PublicDir:     "web/public",
		MainInputFile: "main.go", // Default convention
		AppPort:       "8080",
		Routes:        nil,
		Logger: func(message ...any) {
			// Silent by default
		},
		ExitChan: make(chan bool),
	}
}

func New(c *Config) *ServerHandler {
	// Accept nil and fill missing fields with defaults to avoid panics when caller
	// provides a partially populated Config.
	dc := NewConfig() // default config

	if c == nil {
		c = dc
	} else {
		// Fill zero-value fields with defaults to be defensive
		if c.AppRootDir == "" {
			c.AppRootDir = dc.AppRootDir
		}
		if c.SourceDir == "" {
			c.SourceDir = dc.SourceDir
		}
		if c.OutputDir == "" {
			c.OutputDir = dc.OutputDir
		}
		if c.PublicDir == "" {
			c.PublicDir = dc.PublicDir
		}
		if c.MainInputFile == "" {
			c.MainInputFile = dc.MainInputFile
		}
		if c.AppPort == "" {
			c.AppPort = dc.AppPort
		}
		if c.Logger == nil {
			c.Logger = func(message ...any) {}
		}
		if c.ExitChan == nil {
			c.ExitChan = make(chan bool)
		}
		if c.ArgumentsForCompilingServer == nil {
			c.ArgumentsForCompilingServer = func() []string { return nil }
		}
		if c.ArgumentsToRunServer == nil {
			c.ArgumentsToRunServer = func() []string { return nil }
		}
	}

	sh := &ServerHandler{
		Config:                 c,
		mainFileExternalServer: c.MainInputFile, // Use configured file name
	}

	// Determine initial strategy
	// Check if external server file exists in source directory
	mainFilePath := filepath.Join(c.AppRootDir, c.SourceDir, sh.mainFileExternalServer)
	if _, err := os.Stat(mainFilePath); err == nil {
		// File exists, use External Strategy
		sh.inMemory = false
		sh.strategy = newExternalStrategy(sh)
		sh.Logger("Found existing server file, using External Process strategy.")
	} else {
		// File does not exist, use In-Memory Strategy
		sh.inMemory = true
		sh.strategy = newInMemoryStrategy(sh)
		sh.Logger("No existing server file, defaulting to In-Memory strategy.")
	}

	return sh
}

// MainInputFileRelativePath returns the path relative to AppRootDir (e.g., "src/cmd/appserver/main.go")
func (h *ServerHandler) MainInputFileRelativePath() string {
	return filepath.Join(h.SourceDir, h.mainFileExternalServer)
}

func (h *ServerHandler) SupportedExtensions() []string {
	return []string{".go"}
}

// UnobservedFiles returns the list of files that should not be tracked by file watchers
func (h *ServerHandler) UnobservedFiles() []string {
	if !h.inMemory {
		// If external, we ideally delegate to the external strategy to know what to ignore.
		// But UnobservedFiles is called by the watcher which might be independent.
		// For now, we can check if strategy implements a method or just return standard ignores if known.
		// The previous implementation utilized h.goCompiler.UnobservedFiles().
		// We can cast strategy to externalStrategy or return empty if in-memory.
		if ext, ok := h.strategy.(*externalStrategy); ok {
			return ext.goCompiler.UnobservedFiles()
		}
	}
	// In-memory generally doesn't produce artifacts to ignore, except maybe logs?
	return []string{}
}
