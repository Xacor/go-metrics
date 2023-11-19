package converter

import (
	"github.com/Xacor/go-metrics/internal/server/model"
	pb "github.com/Xacor/go-metrics/proto"
)

func ModelToProto(m model.Metrics) *pb.Metric {
	return &pb.Metric{
		Id:    m.Name,
		Type:  m.MType,
		Delta: *m.Delta,
		Value: *m.Value,
	}
}

func ProtoToModel(p *pb.Metric) model.Metrics {
	value := p.GetValue()
	delta := p.GetDelta()

	return model.Metrics{
		Name:  p.GetId(),
		MType: p.GetType(),
		Value: &value,
		Delta: &delta,
	}
}

func SliceModelToProto(m []model.Metrics) []*pb.Metric {
	res := make([]*pb.Metric, 0, len(m))
	for i := range m {
		res = append(res, ModelToProto(m[i]))
	}

	return res
}

func SliceProtoToModel(p []*pb.Metric) []model.Metrics {
	res := make([]model.Metrics, 0, len(p))
	for i := range p {
		res = append(res, ProtoToModel(p[i]))
	}

	return res
}
