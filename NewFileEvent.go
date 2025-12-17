package server

// event: create,write,remove,rename
func (h *ServerHandler) NewFileEvent(fileName, extension, filePath, event string) error {
	return h.strategy.HandleFileEvent(fileName, extension, filePath, event)
}
