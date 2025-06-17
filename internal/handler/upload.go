package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ankittk/email-audit-service/internal/config"
	"github.com/ankittk/email-audit-service/internal/emailparse"
	"github.com/ankittk/email-audit-service/internal/storage"
	pb "github.com/ankittk/email-audit-service/proto"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Status  string      `json:"status"`
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(code),
		Code:    code,
		Message: message,
	})
}

func respondWithSuccess(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SuccessResponse{
		Data:    data,
		Message: message,
		Status:  "success",
	})
}

func (h *Handler) UploadEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing or invalid file")
		return
	}
	defer file.Close()

	// Validate file type
	if header.Header.Get("Content-Type") != "message/rfc822" &&
		!isEMLFile(header.Filename) {
		respondWithError(w, http.StatusBadRequest, "Only .eml files are supported")
		return
	}

	// Get user and company IDs from form or headers
	userID := r.FormValue("user_id")
	companyID := r.FormValue("company_id")
	if userID == "" || companyID == "" {
		respondWithError(w, http.StatusBadRequest, "user_id and company_id are required")
		return
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(config.GRPCServerAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to create gRPC client: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse the email
	parseResp, err := emailparse.StreamParseEmail(ctx, conn, file)
	if err != nil {
		log.Printf("Email parsing failed: %v", err)
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				respondWithError(w, http.StatusBadRequest, "Invalid email format")
			case codes.DeadlineExceeded:
				respondWithError(w, http.StatusRequestTimeout, "Email processing timeout")
			default:
				respondWithError(w, http.StatusInternalServerError, "Email processing failed")
			}
		} else {
			respondWithError(w, http.StatusInternalServerError, "Email processing failed")
		}
		return
	}

	if parseResp.Status != "SUCCESS" {
		respondWithError(w, http.StatusBadRequest,
			fmt.Sprintf("Email parsing failed: %s", parseResp.ErrorMessage))
		return
	}

	// Get rules
	client := pb.NewRulesEngineServiceClient(conn)
	rulesResp, err := client.ListRules(ctx, &pb.ListRulesRequest{
		CompanyId:   companyID,
		EnabledOnly: true,
	})
	if err != nil {
		log.Printf("Failed to list rules: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve audit rules")
		return
	}

	if len(rulesResp.Rules) == 0 {
		respondWithError(w, http.StatusPreconditionFailed,
			"No audit rules configured for this company")
		return
	}

	// Generate audit ID
	auditID := uuid.New().String()

	// Evaluate rules
	evalResp, err := client.EvaluateRules(ctx, &pb.RulesEvaluationRequest{
		AuditId:     auditID,
		UserId:      userID,
		RuleSet:     rulesResp.Rules,
		EmailThread: parseResp.EmailThread,
	})
	if err != nil {
		log.Printf("Rule evaluation failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Rule evaluation failed")
		return
	}

	// Calculate overall score
	var totalWeight, weightedScore float64
	for _, eval := range evalResp.Evaluations {
		totalWeight += eval.Weight
		weightedScore += eval.Score * eval.Weight
	}

	overallScore := 0.0
	if totalWeight > 0 {
		overallScore = (weightedScore / totalWeight) * 100
	}

	// Create audit report
	auditReport := &pb.AuditReport{
		AuditId:         auditID,
		UserId:          userID,
		CompanyId:       companyID,
		OverallScore:    overallScore,
		RuleEvaluations: evalResp.Evaluations,
		Summary:         generateSummary(evalResp.Evaluations, overallScore),
		CreatedAt:       timestamppb.Now(),
		EmailThread:     parseResp.EmailThread,
	}

	// Store audit report (in-memory for now)
	storage.StoreAuditReport(auditReport)

	respondWithSuccess(w, map[string]interface{}{
		"audit_id":      auditID,
		"overall_score": overallScore,
		"evaluations":   evalResp.Evaluations,
		"summary":       auditReport.Summary,
	}, "Email audit completed successfully")
}

func isEMLFile(filename string) bool {
	return len(filename) > 4 && filename[len(filename)-4:] == ".eml"
}

func generateSummary(evaluations []*pb.RuleEvaluation, overallScore float64) string {
	passed := 0
	failed := 0
	var suggestions []string

	for _, eval := range evaluations {
		if eval.Passed {
			passed++
		} else {
			failed++
			suggestions = append(suggestions, eval.Suggestions...)
		}
	}

	summary := fmt.Sprintf("Overall Score: %.1f%%. %d rules passed, %d rules failed.",
		overallScore, passed, failed)

	if len(suggestions) > 0 {
		summary += " Areas for improvement: " + fmt.Sprintf("%v", suggestions[:min(3, len(suggestions))])
	}

	return summary
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
