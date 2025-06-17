package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/ankittk/email-audit-service/proto"
)

func (h *Handler) ListRulesHandler(w http.ResponseWriter, r *http.Request) {
	companyID := r.URL.Query().Get("company_id")
	category := r.URL.Query().Get("category")
	enabledOnly := r.URL.Query().Get("enabled_only") == "true"

	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.RulesClient.ListRules(ctx, &pb.ListRulesRequest{
		CompanyId:   companyID,
		Category:    category,
		EnabledOnly: enabledOnly,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve rules")
		return
	}

	respondWithSuccess(w, resp.Rules, "Rules retrieved successfully")
}

func (h *Handler) AddRuleHandler(w http.ResponseWriter, r *http.Request) {
	var rule pb.Rule
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	if err := protojson.Unmarshal(body, &rule); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}

	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.RulesClient.AddRule(ctx, &pb.AddRuleRequest{
		Rule:      &rule,
		CompanyId: companyID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add rule: "+err.Error())
		return
	}

	respondWithSuccess(w, map[string]string{
		"rule_id": resp.RuleId,
		"status":  resp.Status,
	}, "Rule added successfully")
}

func (h *Handler) UpdateRuleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["rule_id"]

	var rule pb.Rule
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	if err := protojson.Unmarshal(body, &rule); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format: "+err.Error())
		return
	}
	rule.Id = ruleID

	companyID := r.URL.Query().Get("company_id")
	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.RulesClient.UpdateRule(ctx, &pb.UpdateRuleRequest{
		Rule:      &rule,
		CompanyId: companyID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update rule")
		return
	}

	respondWithSuccess(w, map[string]string{
		"status": resp.Status,
	}, "Rule updated successfully")
}

func (h *Handler) DeleteRuleHandler(w http.ResponseWriter, r *http.Request) {
	ruleID := mux.Vars(r)["rule_id"]
	companyID := r.URL.Query().Get("company_id")

	if companyID == "" {
		respondWithError(w, http.StatusBadRequest, "company_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.RulesClient.DeleteRule(ctx, &pb.DeleteRuleRequest{
		RuleId:    ruleID,
		CompanyId: companyID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete rule")
		return
	}

	respondWithSuccess(w, map[string]string{
		"status": resp.Status,
	}, "Rule deleted successfully")
}
