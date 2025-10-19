package leave

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, leave *LeaveRequest) error
	GetByID(ctx context.Context, id string) (*LeaveRequest, error)
	GetByEmployeeID(ctx context.Context, employeeID string) (*LeaveRequest, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *ListLeaveRequestsRequest) (*ListLeaveRequestsResponse, error)
	Update(ctx context.Context, id string, req *UpdateLeaveRequestRequest) error
	ApproveLeave(ctx context.Context, id string, req *ApproveLeaveRequestRequest) error
	RejectLeave(ctx context.Context, id string, req *RejectLeaveRequestRequest) error
	LeaveBalance(ctx context.Context, req *GetEmployeeLeaveBalanceRequest) (*GetEmployeeLeaveBalanceResponse, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, leave *LeaveRequest) error {
	if err := r.db.WithContext(ctx).Create(leave).Error; err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*LeaveRequest, error) {
	var leaveRequest LeaveRequest
	err := r.db.WithContext(ctx).
		Preload("Approver").
		Preload("Employee").
		Where("id=?", id).
		First(&leaveRequest).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("leave by id (%s) not found", id)
		}
		return nil, fmt.Errorf("failed to get leave by ID %s: %w", id, err)
	}
	return &leaveRequest, nil
}

func (r *repository) GetByEmployeeID(ctx context.Context, employeeID string) (*LeaveRequest, error) {
	var leaveRequest LeaveRequest
	err := r.db.WithContext(ctx).
		Preload("Approver").
		Preload("Employee").
		Where("employee_id=?", employeeID).
		First(&leaveRequest).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("leave by employee ID (%s) not found", employeeID)
		}
		return nil, fmt.Errorf("fialed to get leave by employee ID %s : %w", employeeID, err)
	}
	return &leaveRequest, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ? AND leave_status = 'PENDING'", id).Delete(&LeaveRequest{}).Error; err != nil {
		return fmt.Errorf("failed to delete leave request with id %s : %w", id, err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, req *ListLeaveRequestsRequest) (*ListLeaveRequestsResponse, error) {
	var leavesBalance []*LeaveRequest
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&LeaveRequest{}).Preload("Approver").Preload("Employee")

	if req.EmployeeID != "" {
		query = query.Where("employee_id = ?", req.EmployeeID)
	}
	if req.LeaveType != "" {
		query = query.Where("leave_type = ?", req.LeaveType)
	}
	if req.Status != "" {
		query = query.Where("leave_status = ?", req.Status)
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count leave balances: %w", err)
	}

	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Order("created_at DESC").Find(&leavesBalance).Error; err != nil {
		return nil, fmt.Errorf("failed to list leave balances: %w", err)
	}

	return &ListLeaveRequestsResponse{
		LeaveRequests: leavesBalance,
		TotalCount:    totalCount,
		Page:          req.Page,
		PageSize:      req.PageSize,
	}, nil
}

func (r *repository) Update(ctx context.Context, id string, req *UpdateLeaveRequestRequest) error {
	updateData := map[string]any{}

	if req.LeaveType != "" {
		updateData["leave_type"] = req.LeaveType
	}
	if !req.StartDate.IsZero() {
		updateData["start_date"] = req.StartDate
	}
	if !req.EndDate.IsZero() {
		updateData["end_date"] = req.EndDate
	}
	if req.Reason != "" {
		updateData["reason"] = req.Reason
	}

	if len(updateData) == 0 {
		return fmt.Errorf("no fields to update")
	}

	if err := r.db.WithContext(ctx).
		Model(&LeaveRequest{}).
		Where("id = ?", id).
		Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update leave request: %w", err)
	}

	return nil
}

func (r *repository) ApproveLeave(ctx context.Context, id string, req *ApproveLeaveRequestRequest) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var leave LeaveRequest
		if err := tx.First(&leave, "id = ?", id).Error; err != nil {
			return fmt.Errorf("leave request not found: %w", err)
		}

		if leave.LeaveStatus != "PENDING" {
			return fmt.Errorf("cannot approve leave with status: %s", leave.LeaveStatus)
		}

		now := time.Now()
		if err := tx.Model(&leave).Updates(map[string]any{
			"leave_status": "APPROVED",
			"approver_id":  req.ApproverID,
			"comments":     req.Comments,
			"approved_at":  &now,
		}).Error; err != nil {
			return fmt.Errorf("failed to approve leave: %w", err)
		}

		// Update LeaveBalance
		var balance LeaveBalance
		if err := tx.Where("employee_id = ? AND leave_type = ? AND year = ?", leave.EmployeeID, leave.LeaveType, leave.StartDate.Year()).First(&balance).Error; err != nil {
			return fmt.Errorf("failed to fetch leave balance: %w", err)
		}

		balance.UsedDays += leave.DaysRequested
		balance.TotalDays = max(balance.TotalDays, balance.UsedDays)

		if err := tx.Save(&balance).Error; err != nil {
			return fmt.Errorf("failed to update leave balance: %w", err)
		}

		return nil
	})

	return err
}

func (r *repository) RejectLeave(ctx context.Context, id string, req *RejectLeaveRequestRequest) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var leave LeaveRequest
		if err := tx.First(&leave, "id = ?", id).Error; err != nil {
			return fmt.Errorf("leave request not found: %w", err)
		}

		if leave.LeaveStatus != "PENDING" {
			return fmt.Errorf("cannot reject leave with status: %s", leave.LeaveStatus)
		}

		now := time.Now()
		if err := tx.Model(&leave).Updates(map[string]any{
			"leave_status": "REJECTED",
			"approver_id":  req.ApproverID,
			"comments":     req.Comments,
			"approved_at":  &now,
		}).Error; err != nil {
			return fmt.Errorf("failed to reject leave: %w", err)
		}

		return nil
	})

	return err
}

func (r *repository) LeaveBalance(ctx context.Context, req *GetEmployeeLeaveBalanceRequest) (*GetEmployeeLeaveBalanceResponse, error) {
	query := r.db.WithContext(ctx).Model(&LeaveBalance{}).Where("employee_id = ?", req.EmployeeID)

	if req.Year != 0 {
		query = query.Where("year = ?", req.Year)
	}

	var balances []*LeaveBalance
	if err := query.Find(&balances).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch leave balances: %w", err)
	}

	for _, b := range balances {
		b.RemainingDays = b.GetRemainingDays()
	}

	return &GetEmployeeLeaveBalanceResponse{LeaveBalances: balances}, nil
}
