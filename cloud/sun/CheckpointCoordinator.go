package sun

import "sync"

type CheckpointCoordinator struct {
	mutex sync.Mutex
}
