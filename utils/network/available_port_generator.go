package network

import (
	"net"
	"sync"
)

// Singleton pattern for AvailablePortGenerator
type AvailablePortGenerator struct {
	currPort int
	mu       sync.Mutex
}

var instance *AvailablePortGenerator
var once sync.Once

func GetAvailablePortGeneratorInstance(startPort int) *AvailablePortGenerator {
	once.Do(func() {
		instance = &AvailablePortGenerator{currPort: startPort}
	})
	return instance
}

// Check if a port is available
func (ap *AvailablePortGenerator) portIsAvailable(port int) bool {
	address := "0.0.0.0:" + string(port)
	conn, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Get next available port
func (ap *AvailablePortGenerator) Next() int {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	for !ap.portIsAvailable(ap.currPort) {
		ap.currPort++
	}
	availablePort := ap.currPort
	ap.currPort++
	return availablePort
}
