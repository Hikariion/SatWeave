package operators

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"satweave/messenger/common"
	common2 "satweave/utils/common"
	"satweave/utils/logger"
)

type HttpFFTSource struct {
	JobId        string
	InputChannel chan *common.Record
	nextDataId   uint64
}

// DataStruct 定义了一个可以序列化为 JSON 的结构体
type DataStruct struct {
	DataID  uint64 `json:"dataId"`
	Content []byte `json:"content"`
	Type    string `json:"type"`
}

func (op *HttpFFTSource) Init(initMap map[string]interface{}) {
	op.InputChannel = initMap["InputChannel"].(chan *common.Record)
	// 初始化为 -1
	op.nextDataId = initMap["nextDateId"].(uint64)

	r := gin.Default()

	srv := &http.Server{
		Addr:    ":28899",
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())

	r.POST("/data", func(c *gin.Context) {
		var data DataStruct

		if err := c.ShouldBind(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if data.DataID >= op.nextDataId {
			op.nextDataId = data.DataID + 1
			record := &common.Record{}
			if data.Type == "binary" {
				record.DataType = common.DataType_BINARY
				record.Data = data.Content
			} else if data.Type == "checkpoint" {
				record.DataType = common.DataType_CHECKPOINT
			} else {
				record.DataType = common.DataType_FINISH
			}

			op.InputChannel <- record

			if data.Type == "finish" {
				cancel()
			}
		}
	})

	go func() {
		<-ctx.Done() // 等待取消信号
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Errorf("Failed to shutdown server: %v", err)
		}
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to run server: %v", err)
		}
	}()

}

func (op *HttpFFTSource) Compute(data []byte) ([]byte, error) {
	return nil, nil
}

func (op *HttpFFTSource) SetName(JobId string) {
	op.JobId = JobId
}

func (op *HttpFFTSource) IsSourceOp() bool {
	return true
}

func (op *HttpFFTSource) IsSinkOp() bool {
	return false
}

func (op *HttpFFTSource) IsKeyByOp() bool {
	return false
}

func (op *HttpFFTSource) Checkpoint() []byte {
	return common2.Uint64ToBytes(op.nextDataId)
}

func (op *HttpFFTSource) RestoreFromCheckpoint(data []byte) error {
	return nil
}
