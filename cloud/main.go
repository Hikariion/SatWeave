package cloud

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/logger"
)

var Sun *sun.Sun

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/", hello)
	return router
}

func hello(c *gin.Context) {
	c.String(http.StatusOK, "SatWeave cloud running")
}

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init Sun
	logger.Infof("Start init Sun ...")
	rpcServer := messenger.NewRpcServer(3267)
	s := sun.NewSun(rpcServer)
	go func() {
		err := rpcServer.Run()
		if err != nil {
			logger.Errorf("Sun rpc server run err: %v", err)
		}
	}()
	Sun = s
	logger.Infof("Sun init success")

	// init Gin
	router := NewRouter()
	_ = router.Run(":3268")
}
