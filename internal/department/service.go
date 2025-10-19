package department

import (
	"context"

	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) (*Department, error)
	GetDepartment(ctx context.Context, id string) (*Department, error)
	UpdateDocument(ctx context.Context, id string, req *UpdateDepartmentRequest) (*Department, error)
	DeleteDocument(ctx context.Context, id string) error
	ListDepartments(ctx context.Context, req *ListDepartmentsRequest) (*ListDepartmentsResponse, error)
}

type service struct {
	repo   Repository
	logger *logger.Logger
}

func NewService(repo Repository, logger *logger.Logger) Service {
	return &service{repo: repo, logger: logger.ServiceLogger("department")}
}

func (s *service) CreateDepartment(ctx context.Context, req *CreateDepartmentRequest) (*Department, error) {
	s.logger.Info("Creating new Department", "department_name", req.Name)

	if existingDeap, _ := s.repo.GetByName(ctx, req.Name); existingDeap != nil {
		s.logger.Warn("Department already exists", "name", req.Name)
		return nil, status.Error(codes.AlreadyExists, "Department with this name alreday exists")
	}

	department := FromCreateRequest(req)

	if err := s.repo.Create(ctx, department); err != nil {
		s.logger.Error("Failed to create department", "error", err)
		return nil, status.Error(codes.Internal, "Failed to create department")
	}

	s.logger.Info("Department created successfully", "department_name", department.Name, "id", department.ID)

	deap, err := s.repo.GetByID(ctx, department.ID)
	if err != nil {
		s.logger.Error("Failed to get created employee", "error", err)
		return department, nil // Return the basic employee if we can't get the full one
	}

	return deap, nil
}

func (s *service) GetDepartment(ctx context.Context, id string) (*Department, error) {
	s.logger.Info("Getting department", "id", id)

	department, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get department", "id", id, "error", err)
		return nil, status.Error(codes.NotFound, "Department not found")
	}
	return department, nil
}

func (s *service) DeleteDocument(ctx context.Context, id string) error {
	s.logger.Info("Deleting employee", "id", id)
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete department", "id", id, "error", err)
		return status.Error(codes.Internal, "Failed to delete department")
	}

	s.logger.Info("Department deleted successfully", "id", id)
	return nil
}

func (s *service) UpdateDocument(ctx context.Context, id string, req *UpdateDepartmentRequest) (*Department, error) {
	s.logger.Info("Updating department", "id", id)

	if err := s.repo.Update(ctx, id, req); err != nil {
		s.logger.Error("Failed to update department", "id", id, "error", err)
		return nil, status.Error(codes.Internal, "Failed to update department")
	}

	updatedDepartment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get updated employee", "id", id, "error", err)
		return nil, err
	}

	return updatedDepartment, nil
}

func (s *service) ListDepartments(ctx context.Context, req *ListDepartmentsRequest) (*ListDepartmentsResponse, error) {
	s.logger.Info("Listing department", "page", req.Page, "page_size", req.PageSize, "search", req.Search)

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	response, err := s.repo.List(ctx, req)
	if err != nil {
		s.logger.Error("Failed to list departments", "error", err)
		return nil, status.Error(codes.Internal, "Failed to list departments")
	}

	s.logger.Error("successfully listed departments", "count", len(response.Departments), "total", response.TotalCount)
	return response, nil
}
