package grpc

import (
	"context"
	"reforest/internal/models"
	"reforest/internal/service"
	"reforest/pkg/pb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (h *FinanceHandler) getUserID(ctx context.Context) (uuid.UUID, error) {
	if userIDStr, ok := ctx.Value(userIDKey).(string); ok {
		return uuid.Parse(userIDStr)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get("x-user-id"); len(values) > 0 {
			return uuid.Parse(values[0])
		}
	}

	return uuid.Nil, status.Error(codes.Unauthenticated, "user id not found in context")
}

func (h *FinanceHandler) CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.Transaction, error) {
	tx, err := h.financeService.CreateTransaction(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return mapTransactionToProto(tx), nil
}

func (h *FinanceHandler) TopUpWallet(ctx context.Context, req *pb.TopUpRequest) (*pb.Transaction, error) {
	userID, err := h.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := h.financeService.TopUpWallet(ctx, userID, req.Amount, req.DurationSeconds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return mapTransactionToProto(tx), nil
}

func (h *FinanceHandler) HandleWalletWebhook(ctx context.Context, req *pb.WebhookRequest) (*emptypb.Empty, error) {
	err := h.financeService.HandleWalletWebhook(ctx, req.Event, req.Data)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (h *FinanceHandler) GetBalance(ctx context.Context, _ *emptypb.Empty) (*pb.BalanceResponse, error) {
	userID, err := h.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	balance, err := h.financeService.GetBalance(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BalanceResponse{Balance: balance}, nil
}

func (h *FinanceHandler) GetTransactionHistory(ctx context.Context, _ *emptypb.Empty) (*pb.TransactionList, error) {
	userID, err := h.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	txs, err := h.financeService.GetTransactionHistory(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTxs := make([]*pb.Transaction, len(txs))
	for i, tx := range txs {
		pbTxs[i] = mapTransactionToProto(&tx)
	}

	return &pb.TransactionList{Transactions: pbTxs}, nil
}

func (h *FinanceHandler) CheckPaymentExpiry(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := h.financeService.CheckPaymentExpiry(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func mapTransactionToProto(tx *models.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:     tx.ID.String(),
		UserId: tx.UserID.String(),
		Amount: tx.Amount,
		Type:   tx.Type,
		Status:     tx.Payment.Status,
		PaymentUrl: tx.Payment.PaymentURL,
		ExpiresAt:  timestamppb.New(tx.Payment.ExpiresAt),
		InvoiceId:  tx.ID.String(),
	}
}