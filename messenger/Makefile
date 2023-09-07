GOMODCACHE=$(go env GOMODCACHE)

.PHONY:	clean object tide

clean:
	rm *.pb.go


all:
	protoc --go_out=.. --go-grpc_out=.. messenger/proto/*.proto -I messenger/proto/ -I=/Users/bytedance/go/pkg/mod/go.etcd.io/etcd/raft/v3\@v3.5.2/ -I=/Users/bytedance/go/pkg/mod/github.com/gogo/protobuf\@v1.3.2/ --go_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb --go-grpc_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb

bmac:
	protoc --go_out=.. --go-grpc_out=.. messenger/proto/*.proto -I messenger/proto/ -I=/home/qiutb/go/pkg/mod/go.etcd.io/etcd/raft/v3\@v3.5.2/ -I=/home/qiutb/go/pkg/mod/github.com/gogo/protobuf\@v1.3.2/ --go_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb --go-grpc_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb

wsl:
	protoc --go_out=.. --go-grpc_out=.. messenger/proto/moon.proto -I messenger/proto/ -I=/home/qiutb/go/pkg/mod/go.etcd.io/etcd/raft/v3\@v3.5.2/ -I=/home/qiutb/go/pkg/mod/github.com/gogo/protobuf\@v1.3.2/ --go_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb --go-grpc_opt=Mraftpb/raft.proto=go.etcd.io/etcd/raft/v3/raftpb