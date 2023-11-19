package interceptors

import (
	"net"

	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/internal/server/config"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RegisterUnaryInterceptorChain(cfg config.Config) grpc.ServerOption {
	l := logger.Get()
	var ipNet *net.IPNet
	if cfg.TrustedSubnet != "" {
		_, trustedNet, err := net.ParseCIDR(cfg.TrustedSubnet)
		ipNet = trustedNet
		if err != nil {
			l.Fatal("unable to parse trusted subnet", zap.Error(err))
		}
	}
	return grpc.ChainUnaryInterceptor(
		InitCheckSubnet(ipNet),
		InitVerifySignature(cfg.KeyFile),
		logging.UnaryServerInterceptor(InterceptorLogger(l)),
	)
}
