package core

import (
	"context"

	"github.com/Xacor/go-metrics/internal/server/converter"
	"github.com/Xacor/go-metrics/internal/server/model"
	"github.com/Xacor/go-metrics/internal/server/storage"
	pb "github.com/Xacor/go-metrics/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	repo   storage.MetricRepo
	logger *zap.Logger
}

func NewMetricsServer(repo storage.MetricRepo, logger *zap.Logger) *MetricsServer {
	return &MetricsServer{
		repo:   repo,
		logger: logger,
	}
}

func (s *MetricsServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	result, err := s.repo.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{Metric: converter.ModelToProto(result)}, nil
}

func (s *MetricsServer) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListResponse, error) {
	data, err := s.repo.All(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListResponse{Metrics: converter.SliceModelToProto(data)}, nil
}

func (s *MetricsServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	var result model.Metrics
	if _, err := s.repo.Get(ctx, req.GetMetric().GetId()); err != nil {
		result, err = s.repo.Create(ctx, converter.ProtoToModel(req.GetMetric()))
		if err != nil {
			s.logger.Error(err.Error())
			return nil, err
		}
	} else {
		result, err = s.repo.Update(ctx, converter.ProtoToModel(req.GetMetric()))
		if err != nil {
			s.logger.Error(err.Error())
			return nil, err
		}
	}

	return &pb.UpdateResponse{Result: converter.ModelToProto(result)}, nil
}

func (s *MetricsServer) UpdateList(ctx context.Context, req *pb.UpdateListRequest) (*emptypb.Empty, error) {
	if err := s.repo.UpdateBatch(ctx, converter.SliceProtoToModel(req.GetMetric())); err != nil {
		s.logger.Error("error when updating batch", zap.Error(err), zap.Any("batch", req.GetMetric()))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
