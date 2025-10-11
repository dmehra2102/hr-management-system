package employee

import (
	"context"
	"fmt"

	"github.com/dmehra2102/hr-management-system/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	DeleteEmployee(ctx context.Context, id string) error
	GetEmployee(ctx context.Context, id string) (*Employee, error)
	CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error)
	UpdateEmployee(ctx context.Context, id string, req *UpdateEmployeeRequest) (*Employee, error)
	ListEmployees(ctx context.Context, req *ListEmployeesRequest) (*ListEmployeesResponse, error)
	GetEmployeesByDepartment(ctx context.Context, departmentID string, page, pageSie int) (*ListEmployeesResponse, error)
}

type service struct {
	repo   Repository
	logger *logger.Logger
}

func NewService(repo Repository, logger *logger.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) DeleteEmployee(ctx context.Context, id string) error {
	s.logger.Info("Deleting employee", "id", id)

	employee, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee for deletion", "id", id, "error", err)
		return status.Error(codes.NotFound, "Employee not found")
	}

	// Check if employee is a department manager
	// In a real implementation, you would check if this employee is managing any departments
	// and handle the transition appropriately

	// Soft delete the employee
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete employee", "id", id, "error", err)
		return status.Error(codes.Internal, "Failed to delete employee")
	}

	s.logger.Info("Employee deleted successfully", "id", id, "employee_id", employee.EmployeeID)
	return nil
}

func (s *service) GetEmployee(ctx context.Context, id string) (*Employee, error) {
	s.logger.Info("Getting employee", "id", id)

	employee, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee", "id", id, "error", err)
		return nil, status.Error(codes.NotFound, "Employee not found")
	}

	return employee, nil
}

func (s *service) CreateEmployee(ctx context.Context, req *CreateEmployeeRequest) (*Employee, error) {
	s.logger.Info("Creating new employee", "email", req.Email, "employee_id", req.EmployeeID)

	// Check if employee with same email or employee ID already exists
	if existingEmp, _ := s.repo.GetByEmail(ctx, req.Email); existingEmp != nil {
		s.logger.Warn("Employee with email already exists", "email", req.Email)
		return nil, status.Error(codes.AlreadyExists, "Employee with this email already exists")
	}

	if existingEmp, _ := s.repo.GetByEmployeeID(ctx, req.EmployeeID); existingEmp != nil {
		s.logger.Warn("Employee with employee ID already exists", "employee_id", req.Email)
		return nil, status.Error(codes.AlreadyExists, "Employee with this employee ID already exists")
	}

	// Hash password
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, status.Error(codes.Internal, "Failed to process password")
	}

	employee := FromCreateRequest(req)
	employee.PasswordHash = &hashedPassword

	// Validate Department if provided
	if req.DepartmentID != nil && *req.DepartmentID != "" {
		// TODO : check wheather department exists or not
	}

	if err := s.repo.Create(ctx, employee); err != nil {
		s.logger.Error("Failed to create employee", "error", err)
		return nil, status.Error(codes.Internal, "Failed to create employee")
	}

	s.logger.Info("Employee created successfully", "employee_id", employee.EmployeeID, "id", employee.ID)

	// Get the created employee with relationships
	createdEmployee, err := s.repo.GetByID(ctx, employee.ID)
	if err != nil {
		s.logger.Error("Failed to get created employee", "error", err)
		return employee, nil // Return the basic employee if we can't get the full one
	}

	return createdEmployee, nil
}

func (s *service) UpdateEmployee(ctx context.Context, id string, req *UpdateEmployeeRequest) (*Employee, error) {
	s.logger.Info("Updating employee", "id", id)

	employee, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get employee for update", "id", id, "error", err)
		return nil, status.Error(codes.NotFound, "Employee not found")
	}

	if req.Email != "" && req.Email != employee.Email {
		if existingEmp, _ := s.repo.GetByEmail(ctx, req.Email); existingEmp != nil && existingEmp.ID != id {
			s.logger.Warn("Email already exists for another employee", "email", req.Email)
			return nil, status.Error(codes.AlreadyExists, "Email already exists for another employee")
		}
	}

	employee.ApplyUpdate(req)

	if err := s.repo.Update(ctx, employee); err != nil {
		s.logger.Error("Failed to update employee", "id", id, "error", err)
		return nil, status.Error(codes.Internal, "Failed to update employee")
	}

	s.logger.Info("Emmployee updated successfully", "id", id)

	// Get updated employee with relationships
	updatedEmployee, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get updated employee", "id", id, "error", err)
		return employee, nil // Return the basic employee if we can't get the full one
	}

	return updatedEmployee, nil
}

func (s *service) ListEmployees(ctx context.Context, req *ListEmployeesRequest) (*ListEmployeesResponse, error) {
	s.logger.Info("Listing employees", "page", req.Page, "page_size", req.PageSize, "search", req.Search)

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.Page > 100 {
		req.PageSize = 100
	}

	response, err := s.repo.List(ctx, req)
	if err != nil {
		s.logger.Error("Failed to list employees", "error", err)
		return nil, status.Error(codes.Internal, "Failed to list employees")
	}

	s.logger.Info("Successfully listed employees", "count", len(response.Employees), "total", response.TotalCount)
	return response, nil
}

func (s *service) GetEmployeesByDepartment(ctx context.Context, departmentID string, page, pageSize int) (*ListEmployeesResponse, error) {
	s.logger.Info("Getting employees by department", "department_id", departmentID, "page", page, "page_size", pageSize)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	employees, totalCount, err := s.repo.GetByDepartmentID(ctx, departmentID, page, pageSize)
	if err != nil {
		s.logger.Error("Failed to get employees by department", "department_id", departmentID, "error", err)
		return nil, status.Error(codes.Internal, "Failed to get employees by department")
	}

	response := &ListEmployeesResponse{
		Employees:  employees,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	s.logger.Info("Successfully got employees by department", "department_id", departmentID, "count", len(employees), "total", totalCount)
	return response, nil
}

func (s *service) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}