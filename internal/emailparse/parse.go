package emailparse

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	pb "github.com/ankittk/email-audit-service/proto"
)

func StreamParseEmail(ctx context.Context, conn *grpc.ClientConn, r io.Reader) (*pb.ParseEmailResponse, error) {
	client := pb.NewEmailProcessingServiceClient(conn)
	stream, err := client.ParseEmail(ctx)
	if err != nil {
		return nil, fmt.Errorf("open stream: %w", err)
	}

	buf := make([]byte, 32*1024) // 32KB chunks
	firstChunk := true
	requestID := uuid.New().String()

	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read chunk: %w", err)
		}
		if n == 0 {
			break
		}

		req := &pb.ParseEmailRequest{
			ChunkData: buf[:n],
		}
		if firstChunk {
			req.RequestId = requestID
			firstChunk = false
		}

		if err := stream.Send(req); err != nil {
			return nil, fmt.Errorf("send chunk: %w", err)
		}
		if err == io.EOF {
			break
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, fmt.Errorf("close stream: %w", err)
	}
	return resp, nil
}
