package server

// Store defines the minimal interface for persistent storage
type Store interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

// UI defines the minimal interface for UI interaction
type UI interface {
	RefreshUI()
}

const StoreKeyExternalServer = "server_external_mode"

type ServerModeHandler struct {
	h  *ServerHandler
	db Store
	ui UI
}

func NewServerModeHandler(h *ServerHandler, db Store, ui UI) *ServerModeHandler {
	return &ServerModeHandler{
		h:  h,
		db: db,
		ui: ui,
	}
}

func (s *ServerModeHandler) Name() string {
	return "ServerMode"
}

func (s *ServerModeHandler) Label() string {
	external := false
	if val, err := s.db.Get(StoreKeyExternalServer); err == nil && val == "true" {
		external = true
	}

	if external {
		return "SERVER: EXTERNAL"
	}
	return "SERVER: INTERNAL"
}

func (s *ServerModeHandler) Execute(progress chan<- string) {
	external := "false"
	if val, err := s.db.Get(StoreKeyExternalServer); err == nil && val == "true" {
		external = "true"
	}

	// Toggle
	if external == "true" {
		external = "false"
	} else {
		external = "true"
	}

	s.db.Set(StoreKeyExternalServer, external)

	// Update handler
	isExternal := (external == "true")
	s.h.SetExternalServerMode(isExternal)

	if isExternal {
		progress <- "Switched to External Server Mode"
	} else {
		progress <- "Switched to Internal Server Mode"
	}

	if s.ui != nil {
		s.ui.RefreshUI()
	}
}
