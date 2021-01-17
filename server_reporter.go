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

func newServerReporter(m *ServerMetrics, rpcType grpcType, fullMethod string) *serverReporter {
	r := &serverReporter{
		serverMetrics: m,
		rpcType:       rpcType,
	}
	r.startTime = time.Now()
	r.serviceName, r.methodName = splitMethodName(fullMethod)

	// update metrics
	r.serverMetrics.counterHelper(METRICS_STARTED_EXP_FMT, string(r.rpcType), r.serviceName, r.methodName).Inc()

	return r
}

func (r *serverReporter) ReceivedMessage() {
	r.serverMetrics.counterHelper(METRICS_STREAM_RECV_EXP_FMT, string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) SentMessage() {
	r.serverMetrics.counterHelper(METRICS_STREAM_SENT_EXP_FMT, string(r.rpcType), r.serviceName, r.methodName).Inc()
}

func (r *serverReporter) Handled(code codes.Code) {
	r.serverMetrics.counterHelper(METRICS_STARTED_EXP_FMT, string(r.rpcType), r.serviceName, r.methodName).Inc()
	if r.serverMetrics.enableHistogram {
		r.serverMetrics.histHelper(METRICS_HANDLE_TIME_EXP_FMT, string(r.rpcType), r.serviceName, r.methodName).Update(time.Since(r.startTime).Seconds())
	}
}
