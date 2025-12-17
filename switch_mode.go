package server

import "errors"

// CreateTemplateServer switches from In-Memory to External mode.
// It generates the server files (if not present), compiles, and runs them.
// This implements the transition from "In-Memory" to "Permanent" (External) mode.
func (h *ServerHandler) CreateTemplateServer(progress chan<- string) error {
	if !h.inMemory {
		if progress != nil {
			progress <- "Server is already in external mode."
		}
		return nil
	}

	if progress != nil {
		progress <- "Stopping In-Memory Server..."
	}
	// Stop the current in-memory server
	if err := h.strategy.Stop(); err != nil {
		return errors.Join(errors.New("failed to stop in-memory server"), err)
	}

	if progress != nil {
		progress <- "Generating server files..."
	}
	// Generate the physical files for the server
	if err := h.generateServerFromEmbeddedMarkdown(); err != nil {
		return errors.Join(errors.New("failed to generate server files"), err)
	}

	if progress != nil {
		progress <- "Switching to External Process Strategy..."
	}
	// Switch strategy state
	h.inMemory = false
	h.strategy = newExternalStrategy(h)

	if progress != nil {
		progress <- "Starting External Server..."
	}
	// Start the new external server (compiles and runs)
	// We pass nil for wg because this is a runtime transition, not application startup
	if err := h.strategy.Start(nil); err != nil {
		// If start fails, should we try to revert?
		// For now, return the error.
		return errors.Join(errors.New("failed to start external server"), err)
	}

	if progress != nil {
		progress <- "Server successfully switched to External mode."
	}
	return nil
}
