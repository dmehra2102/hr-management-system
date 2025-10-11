package employee

import (
	"time"

	employeepb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/employee"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type Employee struct {
	ID          string  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EmployeeID  string  `json:"employee_id" gorm:"uniqueIndex;not null"`
	FirstName   string  `json:"first_name" gorm:"not null"`
	LastName    string  `json:"last_name" gorm:"not null"`
	Email       string  `json:"email" gorm:"uniqueIndex:not null"`
	PhoneNumber *string `json:"phone_number,omitempty"`

	// Department Details
	DepartmentID *string     `json:"department_id,omitempty"`
	Department   *Department `json:"department,omitempty" gorm:"foreignKey:DepartmentID"`

	// Job Details
	Position string    `json:"position,omitempty"`
	Salary   float64   `json:"salary" gorm:"default:0"`
	HireDate time.Time `json:"hire_date" gorm:"not null"`
	Status   string    `json:"status" gorm:"default:'ACTIVE';check:status IN ('ACTIVE','INACTIVE', 'TERMINATED', 'ON_LEAVE')"`

	// Address Details
	Street  *string `json:"street,omitempty"`
	City    *string `json:"city,omitempty"`
	State   *string `json:"state,omitempty"`
	ZipCode *string `json:"zip_code,omitempty"`
	Country string  `json:"country" gorm:"default:'IN'"`

	// Authentication
	PasswordHash string     `json:"-" gorm:"column:password_hash"`
	Role         string     `json:"role" gorm:"default:'EMPLOYEE';check:role IN ('ADMIN', 'HR', 'MANAGER', 'EMPLOYEE')"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Department struct {
	ID          string `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (Employee) TableName() string {
	return "employees"
}

type CreateEmployeeRequest struct {
	EmployeeID   string    `json:"employee_id" validate:"required"`
	FirstName    string    `json:"first_name" validate:"required"`
	LastName     string    `json:"last_name" validate:"required"`
	Email        string    `json:"email" validate:"required,email"`
	PhoneNumber  *string   `json:"phone_number,omitempty"`
	DepartmentID *string   `json:"department_id,omitempty"`
	Position     string    `json:"position,omitempty"`
	Salary       float64   `json:"salary" validate:"required"`
	HireDate     time.Time `json:"hire_date" validate:"required"`
	Street       *string   `json:"street,omitempty"`
	City         *string   `json:"city,omitempty"`
	State        *string   `json:"state,omitempty"`
	ZipCode      *string   `json:"zip_code,omitempty"`
	Country      string    `json:"country,omitempty"`
	Password     string    `json:"password" validate:"required,min=6"`
}

type UpdateEmployeeRequest struct {
	FirstName    string  `json:"first_name,omitempty"`
	LastName     string  `json:"last_name,omitempty"`
	Email        string  `json:"email,omitempty" validate:"omitempty,eamil"`
	PhoneNumber  *string `json:"phone_number,omitempty"`
	DepartmentID *string `json:"department_id,omitempty"`
	Position     string  `json:"position,omitempty"`
	Salary       float64 `json:"salary,omitempty" validate:"gte=8"`
	Status       string  `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE TERMINATED ON_LEAVE"`
	Street       *string `json:"street,omitempty"`
	City         *string `json:"city,omitempty"`
	State        *string `json:"state,omitempty"`
	ZipCode      *string `json:"zip_code,omitempty"`
	Country      string  `json:"country,omitempty"`
}

type ListEmployeesRequest struct {
	Page         int    `json:"page" validate:"min=1"`
	PageSize     int    `json:"page_size" validate:"min=1,max=100"`
	Search       string `json:"search,omitempty"`
	DepartmentID string `json:"department_id,omitempty"`
	Status       string `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE TERMINATED ON_LEAVE"`
}

type ListEmployeesResponse struct {
	Employees  []*Employee `json:"employees"`
	TotalCount int64       `json:"total_count"`
	Page       int         `json:"page"`
	PageSie    int         `json:"page_size"`
}

func (e *Employee) ToProto() *employeepb.Employee {
	emp := &employeepb.Employee{
		Id:         e.ID,
		EmployeeId: e.EmployeeID,
		FirstName:  e.FirstName,
		LastName:   e.LastName,
		Email:      e.Email,
		Position:   e.Position,
		Salary:     e.Salary,
		HireDate:   timestamppb.New(e.HireDate),
		CreatedAt:  timestamppb.New(e.CreatedAt),
		UpdatedAt:  timestamppb.New(e.UpdatedAt),
	}

	if e.PhoneNumber != nil {
		emp.PhoneNumber = *e.PhoneNumber
	}
	if e.DepartmentID != nil {
		emp.DepartmentId = *e.DepartmentID
	}

	switch e.Status {
	case "ACTIVE":
		emp.Status = employeepb.EmployeeStatus_EMPLOYEE_STATUS_ACTIVE
	case "INACTIVE":
		emp.Status = employeepb.EmployeeStatus_EMPLOYEE_STATUS_INACTIVE
	case "TERMINATED":
		emp.Status = employeepb.EmployeeStatus_EMPLOYEE_STATUS_TERMINATED
	case "ON_LEAVE":
		emp.Status = employeepb.EmployeeStatus_EMPLOYEE_STATUS_ON_LEAVE
	default:
		emp.Status = employeepb.EmployeeStatus_EMPLOYEE_STATUS_UNSPECIFIED
	}

	emp.Address = &employeepb.Address{
		Country: e.Country,
	}
	if e.Street != nil {
		emp.Address.Street = *e.Street
	}
	if e.City != nil {
		emp.Address.City = *e.City
	}
	if e.State != nil {
		emp.Address.State = *e.State
	}
	if e.ZipCode != nil {
		emp.Address.ZipCode = *e.ZipCode
	}

	return emp
}

func FromCreateRequest(req *CreateEmployeeRequest) *Employee {
	return &Employee{
		EmployeeID:   req.EmployeeID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		DepartmentID: req.DepartmentID,
		Position:     req.Position,
		Salary:       req.Salary,
		HireDate:     req.HireDate,
		Street:       req.Street,
		City:         req.City,
		State:        req.State,
		ZipCode:      req.ZipCode,
		Country:      req.Country,
		Status:       "ACTIVE",
		Role:         "EMPLOYEE",
	}
}

// ApplyUpdate applies UpdateEmployeeRequest to Employee
func (e *Employee) ApplyUpdate(req *UpdateEmployeeRequest) {
	if req.FirstName != "" {
		e.FirstName = req.FirstName
	}
	if req.LastName != "" {
		e.LastName = req.LastName
	}
	if req.Email != "" {
		e.Email = req.Email
	}
	if req.PhoneNumber != nil {
		e.PhoneNumber = req.PhoneNumber
	}
	if req.DepartmentID != nil {
		e.DepartmentID = req.DepartmentID
	}
	if req.Position != "" {
		e.Position = req.Position
	}
	if req.Salary > 0 {
		e.Salary = req.Salary
	}
	if req.Status != "" {
		e.Status = req.Status
	}
	if req.Street != nil {
		e.Street = req.Street
	}
	if req.City != nil {
		e.City = req.City
	}
	if req.State != nil {
		e.State = req.State
	}
	if req.ZipCode != nil {
		e.ZipCode = req.ZipCode
	}
	if req.Country != "" {
		e.Country = req.Country
	}
}

func (e *Employee) FullName() string {
	return e.FirstName + " " + e.LastName
}

func (e *Employee) IsActive() bool {
	return e.Status == "ACTIVE"
}
