package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/config"
	"github.com/Xacor/go-metrics/internal/agent/metric"
	"github.com/Xacor/go-metrics/internal/logger"
	"github.com/Xacor/go-metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"

	poller "github.com/Xacor/go-metrics/internal/agent/http"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printInfo()

	cfg := config.Config{}
	err := cfg.ParseAll()
	if err != nil {
		log.Fatalf("can't parse configuration: %v", err)
	}

	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	l := logger.Get()
	defer l.Sync()

	l.Info("agent configuration", zap.Any("cfg", cfg))

	key, err := cfg.GetKey()
	if err != nil {
		l.Error("failed to get key", zap.Error(err))
	}

	monitor, err := metric.NewMonitor(time.Duration(cfg.GetPollInterval()) * time.Second)
	if err != nil {
		l.Error("failed to create monitor", zap.Error(err))
	}
	defer monitor.Close()

	publicKey, err := cfg.GetPublicKey()
	if err != nil {
		l.Error("failed to get public key", zap.Error(err))
	}

	conn, err := dialGRPC(cfg)
	if err != nil {
		l.Fatal("unable to open client connection", zap.Error(err))
	}
	defer conn.Close()
	metricClient := proto.NewMetricsClient(conn)

	pcfg := poller.PollerConfig{
		ReportInterval: cfg.GetReportInterval(),
		RateLimit:      cfg.GetRateLimit(),
		Address:        cfg.GetURL(),
		Key:            key,
		MetricCh:       monitor.C,
		Client:         &http.Client{},
		GrpcClient:     metricClient,
		Logger:         l,
		PublicKey:      publicKey,
	}

	poller := poller.NewPoller(&pcfg)

	gracefullShutdown := make(chan os.Signal, 2)
	signal.Notify(gracefullShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, stopMonitor := context.WithCancel(context.Background())
	exitCh := make(chan struct{})
	go poller.Run(ctx, exitCh)

	l.Info("signal received, gracefully shutting down", zap.Any("signal", <-gracefullShutdown))
	stopMonitor()
	<-exitCh

	l.Info("gracefully shutting down")
}

func loadTLSCredentials(cafile string) (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(cafile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}

func dialGRPC(cfg config.Config) (*grpc.ClientConn, error) {
	creds, err := loadTLSCredentials(cfg.CACertFile)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(cfg.GRPCAddress, grpc.WithTransportCredentials(creds), grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	if err != nil {
		return nil, err
	}

	return conn, nil
}
