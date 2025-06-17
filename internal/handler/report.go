package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/ankittk/email-audit-service/internal/storage"
)

func (h *Handler) GetAuditReportHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auditID := vars["audit_id"]

	if auditID == "" {
		respondWithError(w, http.StatusBadRequest, "audit_id is required")
		return
	}

	report := storage.GetAuditReport(auditID)
	if report == nil {
		respondWithError(w, http.StatusNotFound, "Audit report not found")
		return
	}

	respondWithSuccess(w, report, "Audit report retrieved successfully")
}

func (h *Handler) GetAuditSummaryHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	companyID := r.URL.Query().Get("company_id")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	if userID == "" && companyID == "" {
		respondWithError(w, http.StatusBadRequest, "user_id or company_id is required")
		return
	}

	page := 1
	pageSize := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	reports, total := storage.GetAuditSummary(userID, companyID, page, pageSize)

	respondWithSuccess(w, map[string]interface{}{
		"reports":     reports,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
	}, "Audit summary retrieved successfully")
}
