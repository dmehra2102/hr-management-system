package department

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, department *Department) error
	GetByID(ctx context.Context, id string) (*Department, error)
	GetByName(ctx context.Context, name string) (*Department, error)
	Update(ctx context.Context, id string, req *UpdateDepartmentRequest) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *ListDepartmentsRequest) (*ListDepartmentsResponse, error)
	Count(ctx context.Context) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, department *Department) error {
	if err := r.db.WithContext(ctx).Create(department).Error; err != nil {
		return fmt.Errorf("failed to create department : %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*Department, error) {
	var department Department
	err := r.db.WithContext(ctx).
		Preload("Manager").Preload("Employees").
		Where("id = ?", id).
		First(&department).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department with id %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get department by ID (%s) : %w", id, err)
	}
	return &department, nil
}

func (r *repository) GetByName(ctx context.Context, name string) (*Department, error) {
	var department Department
	err := r.db.WithContext(ctx).
		Preload("Manager").Preload("Employees").
		Where("name = ?", name).
		First(&department).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department with name %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get department by name (%s) : %w", name, err)
	}
	return &department, nil
}

func (r *repository) Update(ctx context.Context, id string, req *UpdateDepartmentRequest) error {
	updateData := map[string]any{}

	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.Location != "" {
		updateData["location"] = req.Location
	}
	if req.Budget != 0 {
		updateData["budget"] = req.Budget
	}
	if req.ManagerID != nil && *req.ManagerID != "" {
		updateData["manager_id"] = *req.ManagerID
	}
	if req.Name != "" {
		updateData["name"] = req.Name
	}

	if len(updateData) == 0 {
		return fmt.Errorf("no fields to update")
	}

	if err := r.db.WithContext(ctx).
		Model(&Department{}).
		Where("id = ?", id).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update department : %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Department{}).Error; err != nil {
		return fmt.Errorf("failed to delete department : %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, req *ListDepartmentsRequest) (*ListDepartmentsResponse, error) {
	var departments []*Department
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&Department{}).Preload("Manager").Preload("Employees")

	if req.Search != "" {
		searchTerm := "%" + strings.ToLower(req.Search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm,
		)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count departments: %w", err)
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&departments).Error; err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	return &ListDepartmentsResponse{
		Departments: departments,
		TotalCount:  totalCount,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}, nil
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Department{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count department : %w", err)
	}
	return total, nil
}
