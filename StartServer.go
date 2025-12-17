package server

import (
	"sync"
)

// Start initiates the server using the current strategy (In-Memory or External)
func (h *ServerHandler) StartServer(wg *sync.WaitGroup) {
	if err := h.strategy.Start(wg); err != nil {
		h.Logger("StartServer error:", err)
	}
}
