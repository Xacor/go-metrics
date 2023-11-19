package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/Xacor/go-metrics/internal/agent/metric"
	"github.com/Xacor/go-metrics/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type PollerConfig struct {
	MetricCh       <-chan metric.UpdateResult
	Client         *http.Client
	GrpcClient     proto.MetricsClient
	Logger         *zap.Logger
	PublicKey      *rsa.PublicKey
	Address        string
	Key            string
	ReportInterval int
	RateLimit      int
}

type Poller struct {
	metricCh       <-chan metric.UpdateResult
	client         *http.Client
	grpcClient     proto.MetricsClient
	logger         *zap.Logger
	publicKey      *rsa.PublicKey
	address        string
	key            string
	reportInterval int
	rateLimit      int
}

func NewPoller(cfg *PollerConfig) *Poller {
	p := &Poller{
		reportInterval: cfg.ReportInterval,
		address:        cfg.Address,
		metricCh:       cfg.MetricCh,
		client:         cfg.Client,
		grpcClient:     cfg.GrpcClient,
		logger:         cfg.Logger,
		key:            cfg.Key,
		rateLimit:      cfg.RateLimit,
	}

	if cfg.PublicKey != nil {
		p.publicKey = cfg.PublicKey
	}

	return p
}

func (p *Poller) Run(ctx context.Context, exitCh chan struct{}) {
	p.logger.Info("poller started")
	semaphore := NewSemaphore(p.rateLimit)

	t := time.NewTicker(time.Second * time.Duration(p.reportInterval))
	var snap metric.UpdateResult
	for {
		select {
		case snap = <-p.metricCh:
			if snap.Err != nil {
				p.logger.Error("failed to read from UpdateResult", zap.Error(snap.Err))
				continue
			}

		case <-t.C:
			go func(m metric.Metrics) {
				semaphore.Acquire()
				defer semaphore.Release()

				err := p.Send(m)
				if err != nil {
					p.retry(p.Send, m)
				}
			}(snap.Metrics)

		case <-ctx.Done():
			p.logger.Info("sending latest metrics batch")

			semaphore.Acquire()
			defer semaphore.Release()

			err := p.Send(snap.Metrics)
			if err != nil {
				p.retry(p.Send, snap.Metrics)
			}

			exitCh <- struct{}{}
			return
		}
	}
}

func (p *Poller) Send(m metric.Metrics) error {
	p.sendHTTP(m)
	p.sendGRPC(m)
	return nil
}

func (p *Poller) retry(fn func(metric.Metrics) error, arg metric.Metrics) {
	attempts := 0
	var err error
	for i := 1; i < 5; i += 2 {
		time.Sleep(time.Second * time.Duration(i))
		if err = fn(arg); err == nil {
			return
		}
		attempts++
		p.logger.Error("attempt failed", zap.Error(err), zap.Int("attempt #", attempts))
	}
}

func (p *Poller) sendGRPC(m metric.Metrics) error {
	pb, err := m.ToProto()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if p.key != "" {
		json, err := m.MarshalJSON()
		if err != nil {
			return err
		}
		sign, err := Sign(json, p.key)
		if err != nil {
			return err
		}
		metadata.AppendToOutgoingContext(ctx, "HashSHA256", string(sign))

	}

	ip, err := GetLocalIP()
	if err != nil {
		return err
	}
	metadata.AppendToOutgoingContext(ctx, "X-Real-IP", ip)

	_, err = p.grpcClient.UpdateList(ctx, &proto.UpdateListRequest{Metric: pb})
	if err != nil {
		return errors.Wrap(err, "unable to make grpc call")
	}

	return nil
}

func (p *Poller) sendHTTP(m metric.Metrics) error {
	json, err := m.MarshalJSON()
	if err != nil {
		return err
	}

	compressed, err := Compress(json)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(compressed)

	if p.publicKey != nil {
		encryptedBytes, err := rsa.EncryptOAEP(
			sha256.New(),
			rand.Reader,
			p.publicKey,
			compressed,
			nil)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encryptedBytes)
	}

	request, err := http.NewRequest(http.MethodPost, p.address+"/updates/", reader)
	if err != nil {
		return err
	}

	if p.key != "" {
		sign, err := Sign(json, p.key)
		if err != nil {
			return err
		}
		request.Header.Set("HashSHA256", string(sign))
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	ip, err := GetLocalIP()
	if err != nil {
		return err
	}
	request.Header.Set("X-Real-IP", ip)

	resp, err := p.client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
