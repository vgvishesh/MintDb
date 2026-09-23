package main

import (
	pb "MintDb/proto"
	"context"
	"errors"

	"google.golang.org/protobuf/proto"
)

type grpcServer struct {
	pb.UnimplementedMintDbServer
	db Database
}

func NewGrpcServer(db *Database) *grpcServer {
	return &grpcServer{db: *db}
}

func (s *grpcServer) Get(ctx context.Context,request *pb.GetRequest) (*pb.GetResponse, error) {
	value, ok := s.db.Get(request.Key)
	if !ok {
		err := errors.New("key not found")
		return &pb.GetResponse{Value: value, Error: proto.String(err.Error()) }, err
	}
	return &pb.GetResponse{Value: value, Error: nil}, nil	
}


func (s *grpcServer) Set(ctx context.Context,request *pb.SetRequest) (*pb.SetResponse, error) {
	ok := s.db.Set(request.Key, request.Value)
	if !ok {
		err := errors.New("Failed to store the value")
		return &pb.SetResponse{Result: ok, Error: proto.String(err.Error())}, err;
	}
	return &pb.SetResponse{Result: ok, Error: nil}, nil	
}


func (s *grpcServer) Delete(ctx context.Context,request *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	ok := s.db.Delete(request.Key)
	if !ok {
		err := errors.New("failed to delete the key")
		return &pb.DeleteResponse{Result: ok, Error: proto.String(err.Error())}, errors.New("key not found");
	}
	return &pb.DeleteResponse{Result: ok, Error: nil}, nil	
}