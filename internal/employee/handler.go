package employee

import (
	"context"

	employeepb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/employee"
	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handler struct {
	employeepb.UnimplementedEmployeeServiceServer
	service Service
	logger  *logger.Logger
}

func NewHandler(service Service, logger *logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.HandlerLogger("employee"),
	}
}

func (h *Handler) CreateEmployee(ctx context.Context, req *employeepb.CreateEmployeeRequest) (*employeepb.CreateEmployeeResponse, error) {
	h.logger.Info("CreateEmployee called", "employee_id", req.EmployeeId)

	createReq := &CreateEmployeeRequest{
		EmployeeID:   req.EmployeeId,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PhoneNumber:  stringPtr(req.PhoneNumber),
		DepartmentID: stringPtr(req.DepartmentId),
		Position:     req.Position,
		Salary:       req.Salary,
		HireDate:     req.HireDate.AsTime(),
	}

	if req.Address != nil {
		createReq.Street = stringPtr(req.Address.Street)
		createReq.City = stringPtr(req.Address.City)
		createReq.State = stringPtr(req.Address.State)
		createReq.ZipCode = stringPtr(req.Address.ZipCode)
		createReq.Country = req.Address.Country
	}

	employee, err := h.service.CreateEmployee(ctx, createReq)
	if err != nil {
		h.logger.Error("Failed to create employee", "error", err)
		return nil, err
	}

	return &employeepb.CreateEmployeeResponse{
		Employee: employee.ToProto(),
	}, nil
}

func (h *Handler) GetEmployee(ctx context.Context, req *employeepb.GetEmployeeRequest) (*employeepb.GetEmployeeResponse, error) {
	h.logger.Info("GetEmployee called", "id", req.Id)

	employee, err := h.service.GetEmployee(ctx, req.Id)
	if err != nil {
		h.logger.Error("Failed to get employee", "id", req.Id, "error", err)
		return nil, err
	}

	return &employeepb.GetEmployeeResponse{
		Employee: employee.ToProto(),
	}, nil
}

func (h *Handler) UpdateEmployee(ctx context.Context, req *employeepb.UpdateEmployeeRequest) (*employeepb.UpdateEmployeeResponse, error) {
	h.logger.Info("UpdateEmployee called", "id", req.Id)

	updateReq := &UpdateEmployeeRequest{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PhoneNumber:  stringPtr(req.PhoneNumber),
		DepartmentID: stringPtr(req.DepartmentId),
		Position:     req.Position,
		Salary:       req.Salary,
	}

	switch req.Status {
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_ACTIVE:
		updateReq.Status = "ACTIVE"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_INACTIVE:
		updateReq.Status = "INACTIVE"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_TERMINATED:
		updateReq.Status = "TERMINATED"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_ON_LEAVE:
		updateReq.Status = "ON_LEAVE"
	}

	if req.Address != nil {
		updateReq.Street = stringPtr(req.Address.Street)
		updateReq.City = stringPtr(req.Address.City)
		updateReq.State = stringPtr(req.Address.State)
		updateReq.ZipCode = stringPtr(req.Address.ZipCode)
		updateReq.Country = req.Address.Country
	}

	employee, err := h.service.UpdateEmployee(ctx, req.Id, updateReq)
	if err != nil {
		h.logger.Error("Failed to update employee", "id", req.Id, "error", err)
		return nil, err
	}

	return &employeepb.UpdateEmployeeResponse{
		Employee: employee.ToProto(),
	}, nil
}

func (h *Handler) DeleteEmployee(ctx context.Context, req *employeepb.DeleteEmployeeRequest) (*emptypb.Empty, error) {
	h.logger.Info("DeleteEmployee called", "id", req.Id)

	if err := h.service.DeleteEmployee(ctx, req.Id); err != nil {
		h.logger.Error("Failed to delete employee", "id", req.Id, "error", err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *Handler) GetEmployeesByDepartment(ctx context.Context, req *employeepb.GetEmployeesByDepartmentRequest) (*employeepb.ListEmployeesResponse, error) {
	h.logger.Info("GetEmployeesByDepartment called", "department_id", req.DepartmentId)

	response, err := h.service.GetEmployeesByDepartment(ctx, req.DepartmentId, int(req.Page), int(req.PageSize))
	if err != nil {
		h.logger.Error("Failed to get employees by department", "department_id", req.DepartmentId, "error", err)
		return nil, err
	}

	employees := make([]*employeepb.Employee, len(response.Employees))
	for i, employee := range response.Employees {
		employees[i] = employee.ToProto()
	}

	return &employeepb.ListEmployeesResponse{
		Employees:  employees,
		TotalCount: int32(response.TotalCount),
		Page:       int32(response.Page),
		PageSize:   int32(response.PageSize),
	}, nil
}

func (h *Handler) ListEmployees(ctx context.Context, req *employeepb.ListEmployeesRequest) (*employeepb.ListEmployeesResponse, error) {
	h.logger.Info("ListEmployees called", "page", req.Page, "page_size", req.PageSize)

	listReq := &ListEmployeesRequest{
		Page:         int(req.Page),
		PageSize:     int(req.PageSize),
		Search:       req.Search,
		DepartmentID: req.DepartmentId,
	}

	switch req.Status {
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_ACTIVE:
		listReq.Status = "ACTIVE"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_INACTIVE:
		listReq.Status = "INACTIVE"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_TERMINATED:
		listReq.Status = "TERMINATED"
	case employeepb.EmployeeStatus_EMPLOYEE_STATUS_ON_LEAVE:
		listReq.Status = "ON_LEAVE"
	}

	response, err := h.service.ListEmployees(ctx, listReq)
	if err != nil {
		h.logger.Error("Failed to list employees", "error", err)
		return nil, err
	}

	employees := make([]*employeepb.Employee, len(response.Employees))
	for i, employee := range response.Employees {
		employees[i] = employee.ToProto()
	}

	return &employeepb.ListEmployeesResponse{
		Employees:  employees,
		TotalCount: int32(response.TotalCount),
		Page:       int32(response.Page),
		PageSize:   int32(response.PageSize),
	}, nil
}

// stringPtr returns a pointer to a string if not empty, otherwise nil
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
