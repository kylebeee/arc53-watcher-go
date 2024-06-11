package server

func (s *Arc53WatcherServer) routes() {
	s.GET("/", s.handleHealthCheck())
	s.GET("/provider/:key", s.handleGetARC53Data())
	s.GET("/sync/:providerType/:key", s.handleSyncByProviderID())
}
