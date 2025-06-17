package config

import (
	"os"
)

func GRPCServerAddr() string {
	if addr := os.Getenv("GRPC_SERVER_ADDR"); addr != "" {
		return addr
	}
	return "localhost:9090"
}

func RulesEngineAddr() string {
	if addr := os.Getenv("RULES_ENGINE_ADDR"); addr != "" {
		return addr
	}
	return "localhost:9091"
}
