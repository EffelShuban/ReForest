package grpc

import (
	"context"
	"errors"
	"reforest/internal/models"
	"reforest/internal/service"
	"reforest/pkg/pb"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
		CurrentHeightMeters: float32(t.CurrentHeightMeters),
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

func (h *TreeManagementHandler) getUserID(ctx context.Context) (string, error) {
	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		return userID, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found in context")
	}

	values := md.Get("x-user-id")
	if len(values) == 0 || values[0] == "" {
		return "", status.Error(codes.Unauthenticated, "user id not found in metadata")
	}

	return values[0], nil
}


func (h *TreeManagementHandler) CreateSpecies(ctx context.Context, req *pb.Species) (*pb.Species, error) {
	createdSpecies, err := h.treeService.CreateSpecies(ctx, req)
	if err != nil {
		if errors.Is(err, models.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create species")
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
		return nil, status.Error(codes.NotFound, "species not found")
	}
	return mapSpeciesToProto(species), nil
}

func (h *TreeManagementHandler) ListSpecies(ctx context.Context, _ *emptypb.Empty) (*pb.SpeciesList, error) {
	speciesList, err := h.treeService.ListSpecies(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list species")
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
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update species")
	}
	return mapSpeciesToProto(updatedSpecies), nil
}

func (h *TreeManagementHandler) DeleteSpecies(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeleteSpecies(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to delete species")
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
		return nil, status.Error(codes.Internal, "failed to fetch logs")
	}

	pbLogs := make([]*pb.LogEntry, len(logs))
	for i, l := range logs {
		pbLogs[i] = mapLogToProto(l)
	}

	return &pb.LogList{Logs: pbLogs}, nil
}

func (h *TreeManagementHandler) CreateLog(ctx context.Context, req *pb.CreateLogRequest) (*pb.LogEntry, error) {
	adminID, err := h.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	createdLog, err := h.treeService.CreateLog(ctx, req, adminID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create log")
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
		if errors.Is(err, models.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update log")
	}

	return mapLogToProto(updatedLog), nil
}

func (h *TreeManagementHandler) DeleteLog(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format: %v", err)
	}

	if err := h.treeService.DeleteLog(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to delete log")
	}

	return &emptypb.Empty{}, nil
}


func (h *TreeManagementHandler) CreatePlot(ctx context.Context, req *pb.Plot) (*pb.Plot, error) {
	createdPlot, err := h.treeService.CreatePlot(ctx, req)
	if err != nil {
		if errors.Is(err, models.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create plot")
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
		return nil, status.Error(codes.NotFound, "plot not found")
	}
	return mapPlotToProto(plot), nil
}

func (h *TreeManagementHandler) ListPlots(ctx context.Context, _ *emptypb.Empty) (*pb.PlotList, error) {
	plotList, err := h.treeService.ListPlots(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list plots")
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
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update plot")
	}
	return mapPlotToProto(updatedPlot), nil
}

func (h *TreeManagementHandler) DeletePlot(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeletePlot(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to delete plot")
	}
	return &emptypb.Empty{}, nil
}


func (h *TreeManagementHandler) ListTrees(ctx context.Context, _ *emptypb.Empty) (*pb.TreeList, error) {
	treeList, err := h.treeService.ListTrees(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list trees")
	}

	pbTrees := make([]*pb.Tree, len(treeList))
	for i, t := range treeList {
		pbTrees[i] = mapTreeToProto(t)
	}

	return &pb.TreeList{Trees: pbTrees}, nil
}

func (h *TreeManagementHandler) AdoptTree(ctx context.Context, req *pb.AdoptTreeRequest) (*pb.AdoptTreeResponse, error) {
	sponsorID, err := h.getUserID(ctx)
	if err != nil {
		return nil, err
	}
	intent, paymentURL, invoiceID, err := h.treeService.AdoptTree(ctx, req, sponsorID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to adopt tree: %v", err)
	}

	return &pb.AdoptTreeResponse{
		TreeId:     intent.ID.Hex(), // Returning Intent ID as temporary Tree ID
		PaymentUrl: paymentURL,
		InvoiceId:  invoiceID,
	}, nil
}

func (h *TreeManagementHandler) GetTree(ctx context.Context, req *pb.IdRequest) (*pb.Tree, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	tree, err := h.treeService.GetTree(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "tree not found")
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
		if errors.Is(err, models.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to update tree")
	}
	return mapTreeToProto(updatedTree), nil
}

func (h *TreeManagementHandler) DeleteTree(ctx context.Context, req *pb.IdRequest) (*emptypb.Empty, error) {
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	if err := h.treeService.DeleteTree(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to delete tree")
	}
	return &emptypb.Empty{}, nil
}