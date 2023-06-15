package sun

import "sync"

// Sun used to help satellite nodes become a group
type Sun struct {
	Server
	mu sync.Mutex
}

type Server struct {
	// UnimplementedSunServer
}
