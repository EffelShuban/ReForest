package grpc

import (
	"reforest/internal/service"
	"reforest/pkg/pb"
)

type FinanceHandler struct {
	pb.UnimplementedFinanceServiceServer
	financeService service.FinanceService
}

func NewFinanceHandler(financeService service.FinanceService) *FinanceHandler {
	return &FinanceHandler{
		financeService: financeService,
	}
}