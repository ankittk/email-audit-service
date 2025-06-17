package main

import (
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jhillyerd/enmime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ankittk/email-audit-service/proto"
)

type emailProcessingServer struct {
	pb.UnimplementedEmailProcessingServiceServer
}

func (s *emailProcessingServer) ParseEmail(stream pb.EmailProcessingService_ParseEmailServer) error {
	var fullData []byte
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fullData = append(fullData, req.ChunkData...)
	}

	env, err := enmime.ReadEnvelope(strings.NewReader(string(fullData)))
	if err != nil {
		return stream.SendAndClose(&pb.ParseEmailResponse{
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		})
	}

	// Build EmailThread protobuf
	threadID := uuid.New().String()
	thread := &pb.EmailThread{
		ThreadId:     threadID,
		Subject:      env.GetHeader("Subject"),
		Participants: env.GetHeaderValues("From"),
		Messages:     []*pb.EmailMessage{},
	}

	// For demo, only single message parsed
	msgID := uuid.New().String()
	timestamp := time.Now()
	msg := &pb.EmailMessage{
		MessageId: msgID,
		From:      env.GetHeader("From"),
		To:        env.GetHeaderValues("To"),
		Cc:        env.GetHeaderValues("Cc"),
		Subject:   env.GetHeader("Subject"),
		Body:      env.Text,
		Timestamp: timestamppb.New(timestamp),
		IsHtml:    false,
		Headers:   map[string]string{},
	}

	// Add all headers
	for k, v := range env.Root.Header {
		if len(v) > 0 {
			msg.Headers[k] = v[0]
		}
	}

	thread.Messages = append(thread.Messages, msg)

	return stream.SendAndClose(&pb.ParseEmailResponse{
		EmailThread: thread,
		Status:      "SUCCESS",
	})
}

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterEmailProcessingServiceServer(grpcServer, &emailProcessingServer{})

	log.Println("🚀 gRPC server listening on :9090")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
