// Package Monitor
/*
Package Monitor provides a simple interface for monitoring the health of edge cluster node.

Monitor 将节点状态称作 `status`,
卫星轨道集群中的节点会定期将自身 status 发送给集群 moon leader，只有 moon leader 拥有整个集群的完整 state，
*/
package watcher
