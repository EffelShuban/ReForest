package grpc

import (
	"context"
	"reforest/internal/models"
	"reforest/internal/service"
	"reforest/pkg/pb"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TreeManagementHandler struct {
	pb.UnimplementedTreeServiceServer
	treeService service.TreeManagementService
}

func NewTreeManagementHandler(treeService service.TreeManagementService) *TreeManagementHandler {
	return &TreeManagementHandler{
		treeService: treeService,
	}
}

func mapSpeciesToProto(s *models.Species) *pb.Species {
	return &pb.Species{
		Id:              s.ID.Hex(),
		CommonName:      s.CommonName,
		SpaceRequiredM2: float32(s.SpaceRequiredM2),
		Price:           s.Price,
	}
}

func mapPlotToProto(p *models.Plot) *pb.Plot {
	return &pb.Plot{
		Id:               p.ID.Hex(),
		LocationName:     p.LocationName,
		Address:          p.Address,
		AvailableSpaceM2: float32(p.AvailableSpaceM2),
	}
}

func mapTreeToProto(t *models.Tree) *pb.Tree {
	return &pb.Tree{
		Id:                  t.ID.Hex(),
		SponsorId:           t.SponsorID,
		SpeciesId:           t.SpeciesID.Hex(),
		PlotId:              t.PlotID.Hex(),
		CustomName:          t.CustomName,
		InitialHeightMeters: float32(t.InitialHeightMeters),
		TotalFundedLifetime: t.TotalFundedLifetime,
		LastCareDate:        timestamppb.New(t.LastCareDate),
		AdoptedAt:           timestamppb.New(t.AdoptedAt),
	}
}

func mapLogToProto(l *models.LogEntry) *pb.LogEntry {
	return &pb.LogEntry{
		Id:                  l.ID.Hex(),
		AdoptedTreeId:       l.AdoptedTreeID.Hex(),
		AdminId:             l.AdminID,
		CurrentHeightMeters: float32(l.CurrentHeightMeters),
		Activity:            l.Activity,
		Note:                l.Note,
		RecordedAt:          timestamppb.New(l.RecordedAt),
	}
}


func (h *TreeManagementHandler) CreateSpecies(ctx context.Context, req *pb.Species) (*pb.Species, error) {
	createdSpecies, err := h.treeService.CreateSpecies(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create species: %v", err)
	}
	return mapSpeciesToProto(createdSpecies), nil
}

func (h *TreeManagementHandler) GetSpecies(ctx context.Context, req *pb.IdRequest) (*pb.Species, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	species, err := h.treeService.GetSpecies(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "species not found: %v", err)
	}
	return mapSpeciesToProto(species), nil
}

func (h *TreeManagementHandler) ListSpecies(ctx context.Context, _ *emptypb.Empty) (*pb.SpeciesList, error) {
	speciesList, err := h.treeService.ListSpecies(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list species: %v", err)
	}

	pbSpecies := make([]*pb.Species, len(speciesList))
	for i, s := range speciesList {
		pbSpecies[i] = mapSpeciesToProto(s)
	}

	return &pb.SpeciesList{Species: pbSpecies}, nil
}

func (h *TreeManagementHandler) UpdateSpecies(ctx context.Context, req *pb.Species) (*pb.Species, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	updatedSpecies, err := h.treeService.UpdateSpecies(ctx, id, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update species: %v", err)
	}
	return mapSpeciesToProto(updatedSpecies), nil
}

func (h *TreeManagementHandler) DeleteSpecies(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeleteSpecies(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete species: %v", err)
	}
	return &emptypb.Empty{}, nil
}


func (h *TreeManagementHandler) GetTreeLogs(ctx context.Context, req *pb.IdRequest) (*pb.LogList, error) {
	treeID, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tree ID format: %v", err)
	}

	logs, err := h.treeService.GetLogsByTreeID(ctx, treeID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch logs: %v", err)
	}

	pbLogs := make([]*pb.LogEntry, len(logs))
	for i, l := range logs {
		pbLogs[i] = mapLogToProto(l)
	}

	return &pb.LogList{Logs: pbLogs}, nil
}

func (h *TreeManagementHandler) CreateLog(ctx context.Context, req *pb.CreateLogRequest) (*pb.LogEntry, error) {
	// In a real application, you would extract the admin ID from the JWT token in the context.
	// This is typically done in a gRPC interceptor.
	adminID := "placeholder-admin-id" // TODO: Replace with actual admin ID from context

	createdLog, err := h.treeService.CreateLog(ctx, req, adminID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create log: %v", err)
	}

	return mapLogToProto(createdLog), nil
}

func (h *TreeManagementHandler) UpdateLog(ctx context.Context, req *pb.LogEntry) (*pb.LogEntry, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	updatedLog, err := h.treeService.UpdateLog(ctx, id, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update log: %v", err)
	}

	return mapLogToProto(updatedLog), nil
}

func (h *TreeManagementHandler) DeleteLog(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	if err := h.treeService.DeleteLog(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete log: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// --- Plot Handlers ---

func (h *TreeManagementHandler) CreatePlot(ctx context.Context, req *pb.Plot) (*pb.Plot, error) {
	createdPlot, err := h.treeService.CreatePlot(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create plot: %v", err)
	}
	return mapPlotToProto(createdPlot), nil
}

func (h *TreeManagementHandler) GetPlot(ctx context.Context, req *pb.IdRequest) (*pb.Plot, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	plot, err := h.treeService.GetPlot(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "plot not found: %v", err)
	}
	return mapPlotToProto(plot), nil
}

func (h *TreeManagementHandler) ListPlots(ctx context.Context, _ *emptypb.Empty) (*pb.PlotList, error) {
	plotList, err := h.treeService.ListPlots(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list plots: %v", err)
	}

	pbPlots := make([]*pb.Plot, len(plotList))
	for i, p := range plotList {
		pbPlots[i] = mapPlotToProto(p)
	}

	return &pb.PlotList{
		Plots: pbPlots,
	}, nil
}

func (h *TreeManagementHandler) UpdatePlot(ctx context.Context, req *pb.Plot) (*pb.Plot, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	updatedPlot, err := h.treeService.UpdatePlot(ctx, id, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update plot: %v", err)
	}
	return mapPlotToProto(updatedPlot), nil
}

func (h *TreeManagementHandler) DeletePlot(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeletePlot(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete plot: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Tree Handlers ---

func (h *TreeManagementHandler) ListTrees(ctx context.Context, _ *emptypb.Empty) (*pb.TreeList, error) {
	treeList, err := h.treeService.ListTrees(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list trees: %v", err)
	}

	pbTrees := make([]*pb.Tree, len(treeList))
	for i, t := range treeList {
		pbTrees[i] = mapTreeToProto(t)
	}

	return &pb.TreeList{Trees: pbTrees}, nil
}

func (h *TreeManagementHandler) AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest) (*pb.AdoptTreeResponse, error) {
	// In a real application, you would extract the sponsor ID from the JWT token in the context.
	// For now, we'll use a placeholder.
	sponsorID := "placeholder-sponsor-id" // TODO: Replace with actual sponsor ID from context

	tree, err := h.treeService.AdoptTree(ctx, req, sponsorID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to adopt tree: %v", err)
	}

	return &pb.AdoptTreeResponse{
		TreeId: tree.ID.Hex(),
		// PaymentUrl is part of the response but the logic for it might be handled
		// in a separate payment service or based on the sponsor's balance.
		// Leaving it empty for now.
		PaymentUrl: "",
	}, nil
}

func (h *TreeManagementHandler) GetTree(ctx context.Context, req *pb.IdRequest) (*pb.Tree, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	tree, err := h.treeService.GetTree(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tree not found: %v", err)
	}
	return mapTreeToProto(tree), nil
}

func (h *TreeManagementHandler) UpdateTree(ctx context.Context, req *pb.Tree) (*pb.Tree, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	updatedTree, err := h.treeService.UpdateTree(ctx, id, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update tree: %v", err)
	}
	return mapTreeToProto(updatedTree), nil
}

func (h *TreeManagementHandler) DeleteTree(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeleteTree(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete tree: %v", err)
	}
	return &emptypb.Empty{}, nil
}