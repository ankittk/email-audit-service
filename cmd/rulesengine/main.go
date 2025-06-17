package main

import (
	"context"
	"log"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ankittk/email-audit-service/internal/rulesengine"
	pb "github.com/ankittk/email-audit-service/proto"
)

type rulesEngineServer struct {
	pb.UnimplementedRulesEngineServiceServer
	engine *rulesengine.RulesEngine
}

func (s *rulesEngineServer) AddRule(ctx context.Context, req *pb.AddRuleRequest) (*pb.AddRuleResponse, error) {
	rule := req.GetRule()
	companyID := req.GetCompanyId()

	if rule == nil {
		return nil, status.Errorf(codes.InvalidArgument, "rule must be provided")
	}
	if companyID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "company_id must be provided")
	}
	if rule.Id == "" {
		rule.Id = uuid.New().String()
	}
	now := timestamppb.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	s.engine.AddRule(companyID, rule)

	return &pb.AddRuleResponse{
		RuleId: rule.Id,
		Status: "SUCCESS",
	}, nil
}

func (s *rulesEngineServer) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleResponse, error) {
	rule := req.GetRule()
	if rule == nil || rule.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "valid rule must be provided with ID")
	}

	rule.UpdatedAt = timestamppb.Now()
	companyID := req.GetCompanyId()

	if !s.engine.UpdateRule(companyID, rule) {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}

	return &pb.UpdateRuleResponse{
		Status: "SUCCESS",
	}, nil
}

func (s *rulesEngineServer) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleResponse, error) {
	ruleID := req.GetRuleId()
	companyID := req.GetCompanyId()

	if ruleID == "" || companyID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "rule_id and company_id must be provided")
	}

	if !s.engine.DeleteRule(companyID, ruleID) {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}

	return &pb.DeleteRuleResponse{
		Status: "SUCCESS",
	}, nil
}

func (s *rulesEngineServer) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	rules := s.engine.ListRules(req.GetCompanyId(), req.GetCategory(), req.GetEnabledOnly())
	return &pb.ListRulesResponse{
		Rules: rules,
	}, nil
}

func (s *rulesEngineServer) EvaluateRules(ctx context.Context, req *pb.RulesEvaluationRequest) (*pb.RulesEvaluationResponse, error) {
	if req == nil || len(req.RuleSet) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "rule set must be provided")
	}
	if req.EmailThread == nil {
		return nil, status.Errorf(codes.InvalidArgument, "email thread must be provided")
	}

	evals := rulesengine.EvaluateRulesStateless(req.RuleSet, req.EmailThread)

	return &pb.RulesEvaluationResponse{
		AuditId:     req.AuditId,
		Evaluations: evals,
		Status:      "SUCCESS",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	engine := rulesengine.NewRulesEngine()
	initializeDefaultRules(engine, "default-company")

	server := grpc.NewServer()
	pb.RegisterRulesEngineServiceServer(server, &rulesEngineServer{engine: engine})

	log.Println("🚀 Starting Rules Engine gRPC server on port :9091")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve: %v", err)
	}
}

func initializeDefaultRules(engine *rulesengine.RulesEngine, companyID string) {
	now := timestamppb.Now()

	engine.AddRule(companyID, &pb.Rule{
		Id:          uuid.New().String(),
		Name:        "Professional Greeting",
		Description: "Email should start with a professional greeting",
		Category:    "Professionalism",
		Weight:      1.0,
		Type:        pb.RuleType_REGEX_BASED,
		Config: &pb.RuleConfig{
			RegexPattern: `(?i)(dear|hello|hi|good morning|good afternoon)\s+[a-zA-Z]`,
		},
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})

	engine.AddRule(companyID, &pb.Rule{
		Id:          uuid.New().String(),
		Name:        "Professional Closing",
		Description: "Email should end with a professional closing",
		Category:    "Professionalism",
		Weight:      1.0,
		Type:        pb.RuleType_REGEX_BASED,
		Config: &pb.RuleConfig{
			RegexPattern: `(?i)(best regards|sincerely|thank you|thanks|regards|best)\s*,?\s*[a-zA-Z]`,
		},
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})

	engine.AddRule(companyID, &pb.Rule{
		Id:          uuid.New().String(),
		Name:        "Politeness Check",
		Description: "Email should contain polite language",
		Category:    "Courtesy",
		Weight:      0.8,
		Type:        pb.RuleType_KEYWORD_BASED,
		Config: &pb.RuleConfig{
			Keywords: []string{"please", "thank", "appreciate"},
		},
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
