package infos

import (
	"satweave/utils/errno"
	"satweave/utils/logger"
	"sync"
)

type StorageUpdateFunc func(info Information)

type Storage interface {
	Update(info Information) error
	Delete(id string) error
	Get(id string) (Information, error)
	List(prefix string) ([]Information, error)
	GetAll() ([]Information, error)
	GetSnapshot() ([]byte, error)
	RecoverFromSnapshot(snapshot []byte) error

	// SetOnUpdate set a function, it will be called when any info update
	SetOnUpdate(name string, f StorageUpdateFunc)
	CancelOnUpdate(name string)
}

type StorageFactory interface {
	GetStorage(infoType InfoType) Storage
	GetSnapshot() ([]byte, error)
	RecoverFromSnapshot(snapshot []byte) error
	Close()
}

// StorageRegister can auto operate information to corresponding storage.
// It can identify and call storage by infoType.
type StorageRegister struct {
	rwMutex        sync.RWMutex
	storageMap     map[InfoType]Storage
	storageFactory StorageFactory
}

func (register *StorageRegister) Close() {
	register.rwMutex.Lock()
	defer register.rwMutex.Unlock()
	register.storageMap = make(map[InfoType]Storage)
	register.storageFactory.Close()
	logger.Warningf("info storage register closed")
}

// Register while add an info storage into StorageRegister,
// So StorageRegister can process request of this type of info.
func (register *StorageRegister) Register(infoType InfoType, storage Storage) {
	register.rwMutex.Lock()
	defer register.rwMutex.Unlock()
	register.storageMap[infoType] = storage
}

// GetStorage return the registered storage of the infoType.
func (register *StorageRegister) GetStorage(infoType InfoType) Storage {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	storage, ok := register.storageMap[infoType]
	if !ok {
		logger.Fatalf("get storage of type: %v fail, storage not exist", infoType.String())
	}
	return storage
}

// Update will store the Information into corresponding Storage,
// the Storage must Register before.
func (register *StorageRegister) Update(info Information) error {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	if storage, ok := register.storageMap[info.GetInfoType()]; ok {
		return storage.Update(info)
	}
	return errno.InfoTypeNotSupport
}

// Delete will delete the Information from corresponding Storage
// by id, the Storage must Register before.
func (register *StorageRegister) Delete(infoType InfoType, id string) error {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	if storage, ok := register.storageMap[infoType]; ok {
		return storage.Delete(id)
	}
	return errno.InfoTypeNotSupport
}

// Get will return the Information requested from corresponding Storage
// by id, the Storage must Register before.
func (register *StorageRegister) Get(infoType InfoType, id string) (Information, error) {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	if storage, ok := register.storageMap[infoType]; ok {
		return storage.Get(id)
	}
	return InvalidInfo{}, errno.InfoTypeNotSupport
}

// List will return the Information requested from corresponding Storage
// by prefix, the Storage must Register before.
func (register *StorageRegister) List(infoType InfoType, prefix string) ([]Information, error) {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	if storage, ok := register.storageMap[infoType]; ok {
		return storage.List(prefix)
	}
	return nil, errno.InfoTypeNotSupport
}

// GetSnapshot will return a snapshot of all Information from corresponding Storage
func (register *StorageRegister) GetSnapshot() ([]byte, error) {
	register.rwMutex.RLock()
	defer register.rwMutex.RUnlock()
	// TODO: This way to get need to optimize
	return register.storageFactory.GetSnapshot()
}

// RecoverFromSnapshot will recover all Information from corresponding Storage
func (register *StorageRegister) RecoverFromSnapshot(snapshot []byte) error {
	register.rwMutex.Lock()
	defer register.rwMutex.Unlock()
	return register.storageFactory.RecoverFromSnapshot(snapshot)
}

// StorageRegisterBuilder is used to build StorageRegister.
type StorageRegisterBuilder struct {
	register       *StorageRegister
	storageFactory StorageFactory
}

func (builder *StorageRegisterBuilder) getStorage(infoType InfoType) Storage {
	storage := builder.storageFactory.GetStorage(infoType)
	return storage
}

// registerAllStorage will register all storage into StorageRegister.
// 增加 info 信息时，需要在这里注册
func (builder *StorageRegisterBuilder) registerAllStorage() {
	builder.register.Register(InfoType_NODE_INFO, builder.getStorage(InfoType_NODE_INFO))
	builder.register.Register(InfoType_CLUSTER_INFO, builder.getStorage(InfoType_CLUSTER_INFO))
	builder.register.Register(InfoType_USER_INFO, builder.getStorage(InfoType_USER_INFO))
	builder.register.Register(InfoType_BUCKET_INFO, builder.getStorage(InfoType_BUCKET_INFO))
	builder.register.Register(InfoType_VOLUME_INFO, builder.getStorage(InfoType_VOLUME_INFO))
	builder.register.Register(InfoType_TASK_INFO, builder.getStorage(InfoType_TASK_INFO))
}

// GetStorageRegister return a StorageRegister
func (builder *StorageRegisterBuilder) GetStorageRegister() *StorageRegister {
	builder.registerAllStorage()
	return builder.register
}

// NewStorageRegisterBuilder return a StorageRegisterBuilder
func NewStorageRegisterBuilder(factory StorageFactory) *StorageRegisterBuilder {
	return &StorageRegisterBuilder{
		register: &StorageRegister{
			storageMap:     map[InfoType]Storage{},
			storageFactory: factory,
		},
		storageFactory: factory,
	}
}
