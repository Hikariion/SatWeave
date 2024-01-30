package operators

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"satweave/messenger/common"
	utils_common "satweave/utils/common"
	"satweave/utils/logger"
)

type HttpSource struct {
	BeginDataId       uint64
	BinaryDataChannel chan []byte
	name              string
	finished          bool
	currentDataId     uint64
}

// DataStruct 定义了一个可以序列化为 JSON 的结构体
type DataStruct struct {
	DataID   uint64 `json:"dataId"`
	Content  []byte `json:"content"`
	Finished bool   `json:"finished"`
}

func (op *HttpSource) Init(dataId []byte) {
	if dataId == nil {
		op.BeginDataId = 0
	} else {
		op.BeginDataId = utils_common.BytesToUint64(dataId)
	}

	op.BinaryDataChannel = make(chan []byte, 1000)

	// 用 gin 启动一个 Http 服务
	r := gin.Default()

	// 设置一个路由接收数据
	r.POST("/data", func(c *gin.Context) {
		var data DataStruct
		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if data.DataID >= op.BeginDataId {
			op.currentDataId = data.DataID
			op.BinaryDataChannel <- data.Content
		}
		c.Status(http.StatusOK)
	})

	// 在一个新的协程中启动服务，以避免阻塞
	go func() {
		if err := r.Run(":8080"); err != nil {
			logger.Fatalf("Failed to run server: %v", err)
		}
	}()
}

func (op *HttpSource) Compute([]byte) ([]byte, error) {

	if op.finished {
		// 结束
		// TODO 1.30
	}
	content := <-op.BinaryDataChannel

	record := &common.Record{
		DataType: common.DataType_BINARY,
		Data:     content,
	}

	recordBinaryData, err := record.Marshal()
	if err != nil {
		logger.Errorf("marshal record failed: %v", err)
		return nil, err
	}

	return recordBinaryData, nil
}

func (op *HttpSource) SetName(name string) {
	op.name = name
}

func (op *HttpSource) Checkpoint() []byte {
	return nil
}

func (op *HttpSource) RestoreFromCheckpoint([]byte) error {
	return nil
}

func (op *HttpSource) IsSourceOp() bool {
	return true
}

func (op *HttpSource) IsSinkOp() bool {
	return false
}

func (op *HttpSource) IsKeyByOp() bool {
	return false
}
