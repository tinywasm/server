package server

func (h *ServerHandler) RestartServer() error {
	return h.strategy.Restart()
}
