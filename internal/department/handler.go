package department

import (
	"context"

	departmentpb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/department"
	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	departmentpb.UnimplementedDepartmentServiceServer
	service Service
	logger  *logger.Logger
}

func NewHandler(service Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.HandlerLogger("department"),
	}
}

func (h *Handler) CreateDepartment(ctx context.Context, req *departmentpb.CreateDepartmentRequest) (*departmentpb.CreateDepartmentResponse, error) {
	h.logger.Info("CreateDepartment called", "department_name", req.Name)

	createReq := &CreateDepartmentRequest{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   &req.ManagerId,
		Budget:      req.Budget,
		Location:    req.Location,
	}

	department, err := h.service.CreateDepartment(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create department", "error", err)
		return nil, err
	}

	return &departmentpb.CreateDepartmentResponse{
		Department: department.ToProto(),
	}, nil
}

func (h *Handler) GetDepartment(ctx context.Context, req *departmentpb.GetDepartmentRequest) (*departmentpb.GetDepartmentResponse, error) {
	h.logger.Info("GetDepartment called", "department_id", req.Id)

	department, err := h.service.GetDepartment(ctx, req.GetId())
	if err != nil {
		h.logger.Error("Failed to get department", "error", err)
		return nil, err
	}

	return &departmentpb.GetDepartmentResponse{
		Department: department.ToProto(),
	}, nil
}

func (h *Handler) UpdateDepartment(ctx context.Context, req *departmentpb.UpdateDepartmentRequest) (*departmentpb.UpdateDepartmentResponse, error) {
	h.logger.Info("UpdateDepartment called", "department_id", req.Id)

	updateReq := &UpdateDepartmentRequest{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   &req.ManagerId,
		Budget:      req.Budget,
		Location:    req.Location,
	}

	department, err := h.service.UpdateDepartment(ctx, req.GetId(), updateReq)
	if err != nil {
		h.logger.Error("Failed to update department", "error", err)
		return nil, err
	}

	return &departmentpb.UpdateDepartmentResponse{
		Department: department.ToProto(),
	}, nil
}

func (h *Handler) DeleteDepartment(ctx context.Context, req *departmentpb.DeleteDepartmentRequest) (*emptypb.Empty, error) {
	h.logger.Info("DeleteDepartment called", "id", req.Id)

	if err := h.service.DeleteDepartment(ctx, req.Id); err != nil {
		h.logger.Error("Failed to delete department", "id", req.Id, "error", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) ListDepartments(ctx context.Context, req *departmentpb.ListDepartmentsRequest) (*departmentpb.ListDepartmentsResponse, error) {
	h.logger.Info("ListDepartments called", "page", req.Page, "page_size", req.PageSize)

	listReq := &ListDepartmentsRequest{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
		Search:   req.Search,
	}

	response, err := h.service.ListDepartments(ctx, listReq)
	if err != nil {
		h.logger.Error("Failed to list departments", "error", err)
		return nil, err
	}

	departments := make([]*departmentpb.Department, len(response.Departments))
	for i, department := range response.Departments {
		departments[i] = department.ToProto()
	}

	return &departmentpb.ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  int32(response.TotalCount),
		Page:        int32(response.Page),
		PageSize:    int32(response.PageSize),
	}, nil
}
