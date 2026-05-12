package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/new-energy-monitoring/internal/domain/entity"
	"github.com/new-energy-monitoring/internal/domain/repository"
)

type EdgeService interface {
	RegisterNode(ctx context.Context, name, stationID, ipAddress string) (*entity.EdgeNode, error)
	GetNode(ctx context.Context, id string) (*entity.EdgeNode, error)
	ListNodes(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error)
	UpdateNodeConfig(ctx context.Context, id string, configVersion string) error
	DeployModel(ctx context.Context, nodeID string, modelName, version string) error
	GetNodeStatus(ctx context.Context, id string) (*entity.EdgeNode, error)
	TriggerSync(ctx context.Context, id string) error
}

type edgeService struct {
	repo repository.EdgeNodeRepository
}

func NewEdgeService(repo repository.EdgeNodeRepository) EdgeService {
	return &edgeService{repo: repo}
}

func (s *edgeService) RegisterNode(ctx context.Context, name, stationID, ipAddress string) (*entity.EdgeNode, error) {
	node := &entity.EdgeNode{
		ID:        uuid.New().String(),
		Name:      name,
		StationID: stationID,
		IPAddress: ipAddress,
		Status:    entity.EdgeNodeOffline,
	}
	if err := s.repo.Create(ctx, node); err != nil {
		return nil, fmt.Errorf("register node failed: %w", err)
	}
	return node, nil
}

func (s *edgeService) GetNode(ctx context.Context, id string) (*entity.EdgeNode, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *edgeService) ListNodes(ctx context.Context, stationID *string, status *entity.EdgeNodeStatus) ([]*entity.EdgeNode, error) {
	return s.repo.List(ctx, stationID, status)
}

func (s *edgeService) UpdateNodeConfig(ctx context.Context, id string, configVersion string) error {
	node, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	node.ConfigVersion = configVersion
	return s.repo.Update(ctx, node)
}

func (s *edgeService) DeployModel(ctx context.Context, nodeID string, modelName, version string) error {
	node, err := s.repo.GetByID(ctx, nodeID)
	if err != nil {
		return err
	}
	if node.ModelVersions == nil {
		node.ModelVersions = make(entity.MapJSON)
	}
	node.ModelVersions[modelName] = version
	return s.repo.Update(ctx, node)
}

func (s *edgeService) GetNodeStatus(ctx context.Context, id string) (*entity.EdgeNode, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *edgeService) TriggerSync(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	return err
}

func CheckNodeHealth(lastHeartbeat *time.Time, timeout time.Duration) entity.EdgeNodeStatus {
	if lastHeartbeat == nil {
		return entity.EdgeNodeOffline
	}
	if time.Since(*lastHeartbeat) > timeout {
		return entity.EdgeNodeOffline
	}
	return entity.EdgeNodeOnline
}
