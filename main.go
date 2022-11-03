package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cloudnativedaysjp/dreamkast-external-scaler/dreamkast"
	pb "github.com/cloudnativedaysjp/dreamkast-external-scaler/externalscaler"
)

const (
	defaultDesiredReplicas = 1
)

var (
	jst, _ = time.LoadLocation("Asia/Tokyo")
)

type ExternalScaler struct {
	pb.UnimplementedExternalScalerServer

	dkClient dreamkast.Client
}

func (e *ExternalScaler) IsActive(ctx context.Context, scaledObject *pb.ScaledObjectRef) (*pb.IsActiveResponse, error) {
	result, err := e.isActive(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.IsActiveResponse{Result: result}, nil
}

func (e *ExternalScaler) isActive(ctx context.Context) (bool, error) {
	conferences, err := e.dkClient.ListConferences(ctx)
	if err != nil {
		return false, status.Error(codes.Internal, err.Error())
	}

	now := time.Now().In(jst)
	for _, conference := range conferences {
		for _, day := range conference.ConferenceDays {
			d, err := time.Parse("2006-01-02", day.Date)
			if err != nil {
				return false, status.Error(codes.Internal, err.Error())
			}
			if now.Year() == d.Year() && now.YearDay() == d.YearDay() {
				return true, nil
			}
		}
	}
	return false, nil
}

func (e *ExternalScaler) GetMetricSpec(context.Context, *pb.ScaledObjectRef) (*pb.GetMetricSpecResponse, error) {
	return &pb.GetMetricSpecResponse{
		MetricSpecs: []*pb.MetricSpec{{
			MetricName: "dkThreshold",
			TargetSize: 1,
		}},
	}, nil
}

func (e *ExternalScaler) GetMetrics(ctx context.Context, metricRequest *pb.GetMetricsRequest) (*pb.GetMetricsResponse, error) {
	minReplicas := defaultDesiredReplicas

	active, err := e.isActive(ctx)
	if err != nil {
		return nil, err
	} else if active {
		minReplicasStr := metricRequest.ScaledObjectRef.ScalerMetadata["minReplicas"]
		minReplicas, err = strconv.Atoi(minReplicasStr)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "minReplicas must be specified as number")
		}
	}

	return &pb.GetMetricsResponse{
		MetricValues: []*pb.MetricValue{{MetricValue: int64(minReplicas)}},
	}, nil
}

func main() {
	dkClient, err := dreamkast.NewClient("https://event.cloudnativedays.jp/api/")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterExternalScalerServer(s, &ExternalScaler{
		dkClient: dkClient,
	})

	lis, _ := net.Listen("tcp", ":6000")
	fmt.Println("listenting on :6000")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
