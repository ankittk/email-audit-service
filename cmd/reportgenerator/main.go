package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/ankittk/email-audit-service/proto"
)

type reportGenerationServer struct {
	pb.UnimplementedReportGenerationServiceServer
}

func (s *reportGenerationServer) GenerateReport(ctx context.Context, req *pb.GenerateReportRequest) (*pb.GenerateReportResponse, error) {
	if req.GetAuditId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "audit_id must be provided")
	}

	if req.GetReportData() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "report_data must be provided")
	}

	format := req.GetFormat()
	if format == "" {
		format = "json"
	}

	var content []byte
	var contentType string
	var err error

	switch format {
	case "json":
		content, err = generateJSONReport(req.GetReportData())
		contentType = "application/json"
	case "html":
		content, err = generateHTMLReport(req.GetReportData())
		contentType = "text/html"
	case "pdf":
		return nil, status.Errorf(codes.Unimplemented, "PDF generation not implemented yet")
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported format: %s", format)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate report: %v", err)
	}

	return &pb.GenerateReportResponse{
		ReportContent: content,
		ContentType:   contentType,
		Status:        "SUCCESS",
	}, nil
}

func generateJSONReport(report *pb.AuditReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func generateHTMLReport(report *pb.AuditReport) ([]byte, error) {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Email Audit Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background-color: #f8f9fa; padding: 20px; border-radius: 8px; }
        .score { font-size: 24px; font-weight: bold; color: {{.ScoreColor}}; }
        .evaluation { margin: 20px 0; padding: 15px; border-left: 4px solid {{.BorderColor}}; }
        .passed { border-color: #28a745; background-color: #d4edda; }
        .failed { border-color: #dc3545; background-color: #f8d7da; }
        .email-content { background-color: #f8f9fa; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Email Audit Report</h1>
        <p><strong>Audit ID:</strong> {{.AuditId}}</p>
        <p><strong>User ID:</strong> {{.UserId}}</p>
        <p><strong>Overall Score:</strong> <span class="score">{{printf "%.1f" .OverallScore}}%</span></p>
        <p><strong>Generated:</strong> {{.CreatedAt.AsTime.Format "2006-01-02 15:04:05"}}</p>
    </div>

    <h2>Summary</h2>
    <p>{{.Summary}}</p>

    <h2>Rule Evaluations</h2>
    {{range .RuleEvaluations}}
    <div class="evaluation {{if .Passed}}passed{{else}}failed{{end}}">
        <h3>{{.RuleName}} ({{.Category}}) - {{if .Passed}}✅ PASSED{{else}}❌ FAILED{{end}}</h3>
        <p><strong>Weight:</strong> {{.Weight}} | <strong>Score:</strong> {{printf "%.1f" .Score}}</p>
        {{if .Suggestions}}
        <p><strong>Suggestions:</strong></p>
        <ul>
        {{range .Suggestions}}
            <li>{{.}}</li>
        {{end}}
        </ul>
        {{end}}
    </div>
    {{end}}

    <h2>Email Content</h2>
    {{if .EmailThread}}
    <div class="email-content">
        <h3>Subject: {{.EmailThread.Subject}}</h3>
        {{range .EmailThread.Messages}}
        <div style="margin: 15px 0; padding: 10px; border: 1px solid #ddd;">
            <p><strong>From:</strong> {{.From}}</p>
            <p><strong>To:</strong> {{range .To}}{{.}} {{end}}</p>
            <p><strong>Date:</strong> {{.Timestamp.AsTime.Format "2006-01-02 15:04:05"}}</p>
            <div style="margin-top: 10px; padding: 10px; background-color: white;">
                {{.Body}}
            </div>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	t, err := template.New("report").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, report)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func main() {
	lis, err := net.Listen("tcp", ":9092")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterReportGenerationServiceServer(server, &reportGenerationServer{})

	log.Println("🚀 Report Generation gRPC server is running on port :9092")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}
