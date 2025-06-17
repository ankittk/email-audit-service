package handler

import (
	"net/http"

	"github.com/gorilla/mux"

	pb "github.com/ankittk/email-audit-service/proto"
)

type Handler struct {
	RulesClient pb.RulesEngineServiceClient
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/health", h.HealthCheckHandler).Methods("GET")
	r.HandleFunc("/upload-email", h.UploadEmailHandler).Methods("POST")

	r.HandleFunc("/audit/{audit_id}", h.GetAuditReportHandler).Methods("GET")
	r.HandleFunc("/audit/summary", h.GetAuditSummaryHandler).Methods("GET")

	r.HandleFunc("/rules", h.ListRulesHandler).Methods("GET")
	r.HandleFunc("/rules", h.AddRuleHandler).Methods("POST")
	r.HandleFunc("/rules/{rule_id}", h.UpdateRuleHandler).Methods("PUT")
	r.HandleFunc("/rules/{rule_id}", h.DeleteRuleHandler).Methods("DELETE")
}

func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondWithSuccess(w, map[string]string{"status": "healthy"}, "Service is running")
}
