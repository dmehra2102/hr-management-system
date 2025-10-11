package employee

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, employee *Employee) error
	GetByID(ctx context.Context, id string) (*Employee, error)
	GetByEmail(ctx context.Context, email string) (*Employee, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*Employee, error)
	Update(ctx context.Context, employee *Employee) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *ListEmployeesRequest) (*ListEmployeesResponse, error)
	GetByDepartmentID(ctx context.Context, departmentID string, page, pageSize int) ([]*Employee, int64, error)
	UpdateLastLogin(ctx context.Context, id string, lastLogin time.Time) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	Count(ctx context.Context) (int64, error)
	GetManagers(ctx context.Context) ([]*Employee, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, employee *Employee) error {
	if err := r.db.WithContext(ctx).Create(employee).Error; err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Employee, error) {
	var employee Employee
	err := r.db.WithContext(ctx).
		Preload("Department").
		Where("id=?", id).
		First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by ID: %w", err)
	}
	return &employee, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*Employee, error) {
	var employee Employee
	err := r.db.WithContext(ctx).
		Preload("Department").
		Where("email=?", email).
		First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by email: %w", err)
	}
	return &employee, nil
}

func (r *repository) GetByEmployeeID(ctx context.Context, employeeID string) (*Employee, error) {
	var employee Employee
	err := r.db.WithContext(ctx).
		Preload("Department").
		Where("employee_id = ?", employeeID).
		First(&employee).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by employee ID: %w", err)
	}
	return &employee, nil
}

func (r *repository) Update(ctx context.Context, employee *Employee) error {
	if err := r.db.WithContext(ctx).Save(employee).Error; err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Employee{}).Error; err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, req *ListEmployeesRequest) (*ListEmployeesResponse, error) {
	var employees []*Employee
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&Employee{}).Preload("Department")

	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where(
			"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(employee_id) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	if req.DepartmentID != "" {
		query = query.Where("department_id = ?", req.DepartmentID)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	// total count
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count employees: %w", err)
	}

	// Apply pagination
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	return &ListEmployeesResponse{
		Employees:  employees,
		TotalCount: totalCount,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

func (r *repository) GetByDepartmentID(ctx context.Context, departmentID string, page, pageSize int) ([]*Employee, int64, error) {
	var employees []*Employee
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&Employee{}).
		Where("department_id = ? AND status = ?", departmentID, "ACTIVE").
		Preload("department")

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count employees by department: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get employees by department: %w", err)
	}

	return employees, totalCount, nil
}

func (r *repository) UpdateLastLogin(ctx context.Context, id string, lastLogin time.Time) error {
	if err := r.db.WithContext(ctx).Model(&Employee{}).Where("id = ?", id).Update("last_login_at", lastLogin).Error; err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	if err := r.db.WithContext(ctx).Model(&Employee{}).Where("id = ?", id).Update("password_hash", passwordHash).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

// Count returns the total number of employees
func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Employee{}).Where("status != ?", "TERMINATED").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count employees: %w", err)
	}
	return count, nil
}

func (r *repository) GetManagers(ctx context.Context) ([]*Employee, error) {
	var employees []*Employee
	err := r.db.WithContext(ctx).
		Where("role IN ?", []string{"MANAGER", "HR", "ADMIN"}).
		Where("status = ?", "ACTIVE").
		Order("first_name, last_name").
		Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get managers: %w", err)
	}

	return employees, nil
}
