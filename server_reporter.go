package grpc_victoriametrics

import (
	"google.golang.org/grpc/codes"
	"time"
)

type serverReporter struct {
	serverMetrics *ServerMetrics
	rpcType       grpcType
	serviceName   string
	methodName    string
	startTime     time.Time
}
