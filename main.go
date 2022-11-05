package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
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

	defaultDkClient dreamkast.Client
}

func (e *ExternalScaler) IsActive(ctx context.Context, scaledObject *pb.ScaledObjectRef) (*pb.IsActiveResponse, error) {
	dkEndpointUrl := scaledObject.ScalerMetadata["dkEndpointUrl"]
	var dkClient dreamkast.Client
	if dkEndpointUrl != "" {
		var err error
		dkClient, err = dreamkast.NewClient(dkEndpointUrl)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to connect to specified dkEndpointUrl")
		}
	}
	result, err := e.isActive(ctx, dkClient)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.IsActiveResponse{Result: result}, nil
}

func (e *ExternalScaler) isActive(ctx context.Context, dkClient dreamkast.Client) (bool, error) {
	if dkClient == nil {
		dkClient = e.defaultDkClient
	}
	conferences, err := dkClient.ListConferences(ctx)
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
	dkEndpointUrl := metricRequest.ScaledObjectRef.ScalerMetadata["dkEndpointUrl"]
	var dkClient dreamkast.Client
	if dkEndpointUrl != "" {
		var err error
		dkClient, err = dreamkast.NewClient(dkEndpointUrl)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "failed to connect to specified dkEndpointUrl")
		}
	}

	minReplicas := defaultDesiredReplicas
	active, err := e.isActive(ctx, dkClient)
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

func getenvOrDefault(key, defaultV string) string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultV
	}
	return v
}

func main() {
	dkUrl := getenvOrDefault("DK_ENDPOINT_URL", "https://event.cloudnativedays.jp/")
	dkClient, err := dreamkast.NewClient(dkUrl)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterExternalScalerServer(s, &ExternalScaler{
		defaultDkClient: dkClient,
	})

	lis, _ := net.Listen("tcp", ":6000")
	fmt.Println("listenting on :6000")
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
