package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"satweave/cloud/sun"
	"satweave/messenger"
	"satweave/utils/logger"
	"time"
)

var Sun *sun.Sun

func NewRouter() *gin.Engine {
	router := gin.Default()

	// 配置CORS中间件
	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(config))

	router.GET("/", hello)
	router.POST("/submit-stream-job", SubmitStreamJob)
	router.GET("/node-status", GetNodeStatus)
	return router
}

func GetNodeStatus(c *gin.Context) {
	nodes := Sun.NodeStatusCollector.Nodes
	c.JSON(http.StatusOK, nodes)
}

func hello(c *gin.Context) {
	c.String(http.StatusOK, "SatWeave cloud running")
}

func SubmitStreamJob(c *gin.Context) {
	var request sun.SubmitJobRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果有错误，返回一个400错误和错误消息
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.JobId == "" || request.SatelliteName == "" || request.YamlStr == "" || request.PathNodes == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JobId, SatelliteName, YamlByte, PathNodes must be provided"})
		return
	}

	_, err := Sun.SubmitJob(context.Background(), &request)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func main() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init Sun
	logger.Infof("Start init Sun ...")
	rpcServer := messenger.NewRpcServer(3267)
	Sun = sun.NewSun(rpcServer)
	go func() {
		err := rpcServer.Run()
		if err != nil {
			logger.Errorf("Sun rpc server run err: %v", err)
		}
	}()
	logger.Infof("Sun init success")

	// init Gin
	router := NewRouter()
	_ = router.Run(":3268")
}
