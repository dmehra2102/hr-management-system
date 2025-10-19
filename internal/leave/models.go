package leave

import (
	"fmt"
	"time"

	leavepb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/leave"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type LeaveRequest struct {
	ID            string     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID    string     `json:"employee_id" gorm:"not null;index"`
	Employee      *Employee  `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	LeaveType     string     `json:"leave_type" gorm:"not null;check:leave_type IN ('ANNUAL','SICK','MATERNITY','PATERNITY', 'EMERGENCY', 'PERSONAL')"`
	StartDate     time.Time  `json:"start_date" gorm:"not null"`
	EndDate       time.Time  `json:"end_date" gorm:"not null"`
	DaysRequested int        `json:"days_requested" gorm:"not null"`
	Reason        string     `json:"reason"`
	LeaveStatus   string     `json:"leave_status" gorm:"default:'PENDING';check:leave_status IN ('PENDING','APPROVED','REJECTED','CANCELLED')"`
	ApproverID    *string    `json:"approver_id,omitempty"`
	Approver      *Employee  `json:"approver,omitempty" gorm:"foreignKey:ApproverID"`
	Comments      string     `json:"comments"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type LeaveBalance struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID    string    `json:"employee_id" gorm:"not null;index"`
	Employee      *Employee `json:"employee,omitempty" gorm:"foreignKey:EmployeeID"`
	LeaveType     string    `json:"leave_type" gorm:"not null;check:leave_type IN ('ANNUAL','SICK','MATERNITY', 'PATERNITY', 'EMERGENCY', 'PERSONAL')"`
	Year          int       `json:"year" gorm:"not null"`
	TotalDays     int       `json:"total_days" gorm:"default:0"`
	UsedDays      int       `json:"used_days" gorm:"default:0"`
	RemainingDays int       `json:"remaining_days" gorm:"-"` // Computed field

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Employee struct {
	ID         string `json:"id" gorm:"type:uuid;primaryKey"`
	EmployeeID string `json:"employee_id"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}

func (LeaveRequest) TableName() string {
	return "leaves"
}

func (LeaveBalance) TableName() string {
	return "leave_balances"
}

type CreateLeaveRequestRequest struct {
	EmployeeID string    `json:"employee_id" validate:"required"`
	LeaveType  string    `json:"leave_type" validate:"required,oneof=ANNUAL SICK MATERNITY PATERNITY EMERGENCY PERSONAL"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
	Reason     string    `json:"reason,omitempty"`
}

type UpdateLeaveRequestRequest struct {
	LeaveType string    `json:"leave_type,omitempty" validate:"omitempty,oneof=ANNUAL SICK MATERNITY PATERNITY EMERGENCY PERSONAL"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

type ListLeaveRequestsRequest struct {
	Page       int    `json:"page" validate:"min=1"`
	PageSize   int    `json:"page_size" validate:"min=1,max=100"`
	EmployeeID string `json:"employee_id,omitempty"`
	Status     string `json:"status,omitempty" validate:"omitempty,oneof=PENDING APPROVED REJECTED CANCELLED"`
	LeaveType  string `json:"leave_type,omitempty" validate:"omitempty, oneof=ANNUAL SICK MATERNITY PATERNITY EMERGENCY PERSONAL"`
}

type ListLeaveRequestsResponse struct {
	LeaveRequests []*LeaveRequest `json:"leave_requests"`
	TotalCount    int64           `json:"total_count"`
	Page          int             `json:"page"`
	PageSize      int             `json:"page_size"`
}

type ApproveLeaveRequestRequest struct {
	ApproverID string `json:"approver_id" validate:"required"`
	Comments   string `json:"comments,omitempty"`
}

type RejectLeaveRequestRequest struct {
	ApproverID string `json:"approver_id" validate:"required"`
	Comments   string `json:"comments" validate:"required"`
}

type GetEmployeeLeaveBalanceRequest struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	Year       int    `json:"year,omitempty"`
}

type GetEmployeeLeaveBalanceResponse struct {
	LeaveBalances []*LeaveBalance `json:"leave_balances"`
}

func (lr *LeaveRequest) ToProto() *leavepb.LeaveRequest {
	leave := &leavepb.LeaveRequest{
		Id:            lr.ID,
		EmployeeId:    lr.EmployeeID,
		StartDate:     timestamppb.New(lr.StartDate),
		EndDate:       timestamppb.New(lr.EndDate),
		DaysRequested: int32(lr.DaysRequested),
		Reason:        lr.Reason,
		Comments:      lr.Comments,
		CreatedAt:     timestamppb.New(lr.CreatedAt),
		UpdatedAt:     timestamppb.New(lr.UpdatedAt),
	}

	// Handle employee name
	if lr.Employee != nil {
		leave.EmployeeName = lr.Employee.FirstName + " " + lr.Employee.LastName
	}

	// Handle approver
	if lr.ApproverID != nil {
		leave.ApproverId = *lr.ApproverID
		if lr.Approver != nil {
			leave.ApproverName = lr.Approver.FirstName + " " + lr.Approver.LastName
		}
	}

	// Handle approved at
	if lr.ApprovedAt != nil {
		leave.ApprovedAt = timestamppb.New(*lr.ApprovedAt)
	}

	// Set leave type
	switch lr.LeaveType {
	case "ANNUAL":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_ANNUAL
	case "SICK":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_SICK
	case "MATERNITY":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_MATERNITY
	case "PATERNITY":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_PATERNITY
	case "EMERGENCY":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_EMERGENCY
	case "PERSONAL":
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_PERSONAL
	default:
		leave.LeaveType = leavepb.LeaveType_LEAVE_TYPE_UNSPECIFIED
	}

	// Set status
	switch lr.LeaveStatus {
	case "PENDING":
		leave.LeaveStatus = leavepb.LeaveStatus_LEAVE_STATUS_PENDING
	case "APPROVED":
		leave.LeaveStatus = leavepb.LeaveStatus_LEAVE_STATUS_APPROVED
	case "REJECTED":
		leave.LeaveStatus = leavepb.LeaveStatus_LEAVE_STATUS_REJECTED
	case "CANCELLED":
		leave.LeaveStatus = leavepb.LeaveStatus_LEAVE_STATUS_CANCELLED
	default:
		leave.LeaveStatus = leavepb.LeaveStatus_LEAVE_STATUS_UNSPECIFIED
	}

	return leave
}

// ToProto converts LeaveBalance to protobuf message
func (lb *LeaveBalance) ToProto() *leavepb.LeaveBalance {
	balance := &leavepb.LeaveBalance{
		EmployeeId:    lb.EmployeeID,
		TotalDays:     int32(lb.TotalDays),
		UsedDays:      int32(lb.UsedDays),
		RemainingDays: int32(lb.GetRemainingDays()),
		Year:          int32(lb.Year),
	}

	// Set leave type
	switch lb.LeaveType {
	case "ANNUAL":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_ANNUAL
	case "SICK":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_SICK
	case "MATERNITY":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_MATERNITY
	case "PATERNITY":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_PATERNITY
	case "EMERGENCY":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_EMERGENCY
	case "PERSONAL":
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_PERSONAL
	default:
		balance.LeaveType = leavepb.LeaveType_LEAVE_TYPE_UNSPECIFIED
	}

	return balance
}

func FromCreateRequest(req *CreateLeaveRequestRequest) *LeaveRequest {
	daysRequested := calculateDaysRequested(req.StartDate, req.EndDate)

	return &LeaveRequest{
		EmployeeID:    req.EmployeeID,
		LeaveType:     req.LeaveType,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		DaysRequested: daysRequested,
		Reason:        req.Reason,
		LeaveStatus:   "PENDING",
	}
}

func (lr *LeaveRequest) ApplyUpdate(req *UpdateLeaveRequestRequest) {
	if req.LeaveType != "" {
		lr.LeaveType = req.LeaveType
	}
	if !req.StartDate.IsZero() {
		lr.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		lr.EndDate = req.EndDate
	}
	if req.Reason != "" {
		lr.Reason = req.Reason
	}

	// Recalculate days if dates changed
	if !req.StartDate.IsZero() || !req.EndDate.IsZero() {
		lr.DaysRequested = calculateDaysRequested(lr.StartDate, lr.EndDate)
	}
}

// GetRemainingDays calculates and returns remaining days
func (lb *LeaveBalance) GetRemainingDays() int {
	remaining := lb.TotalDays - lb.UsedDays
	if remaining < 0 {
		return 0
	}
	return remaining
}

// HasSufficientBalance checks if there are sufficient days for a request
func (lb *LeaveBalance) HasSufficientBalance(requestedDays int) bool {
	return lb.GetRemainingDays() >= requestedDays
}

// calculateDaysRequested calculates the number of days between start and end date
func calculateDaysRequested(startDate, endDate time.Time) int {
	// Simple calculation - in a real system, you'd exclude weekends and holidays
	duration := endDate.Sub(startDate)
	days := int(duration.Hours()/24) + 1 // Include both start and end dates
	if days < 0 {
		return 0
	}
	return days
}

func (lr *LeaveRequest) GetEmployeeName() string {
	if lr.Employee != nil {
		return lr.Employee.FirstName + " " + lr.Employee.LastName
	}
	return ""
}

func (lr *LeaveRequest) GetApproverName() string {
	if lr.Approver != nil {
		return lr.Approver.FirstName + " " + lr.Approver.LastName
	}
	return ""
}

// GetDuration returns the duration of the leave in a human-readable format
func (lr *LeaveRequest) GetDuration() string {
	if lr.DaysRequested == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", lr.DaysRequested)
}
