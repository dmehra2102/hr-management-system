package department

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	departmentpb "github.com/dmehra2102/hr-management-system/api/proto/v1/gen/department"
)

type Department struct {
	ID            string    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"not null"`
	Description   string    `json:"description"`
	ManagerID     *string   `json:"manager_id,omitempty"`
	Manager       *Employee `json:"manager,omitempty" gorm:"foreignKey:ManagerID"`
	Budget        float64   `json:"budget" gorm:"default:0"`
	Location      string    `json:"location"`
	EmployeeCount int       `json:"employee_count" gorm:"default:0"`

	Employees []Employee `json:"employees,omitempty" gorm:"foreignKey:DepartmentID"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type Employee struct {
	ID           string  `json:"id" gorm:"type:uuid;primaryKey"`
	EmployeeID   string  `json:"employee_id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	DepartmentID *string `json:"department_id,omitempty"`
	Position     string  `json:"position"`
	Status       string  `json:"status"`
}

func (Department) TableName() string {
	return "departments"
}

type CreateDepartmentRequest struct {
	Name        string  `json:"name" validate:"required,max=100"`
	Description string  `json:"description,omitempty"`
	ManagerID   *string `json:"manager_id,omitempty"`
	Budget      float64 `json:"budget" validate:"gte=0"`
	Location    string  `json:"location,omitempty"`
}

type UpdateDepartmentRequest struct {
	Name        string  `json:"name,omitempty" validate:"omitempty,max=100"`
	Description string  `json:"description,omitempty"`
	ManagerID   *string `json:"manager_id,omitempty"`
	Budget      float64 `json:"budget,omitempty" validate:"gte=0"`
	Location    string  `json:"location,omitempty"`
}

type ListDepartmentsRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Search   string `json:"search,omitempty"`
}

type ListDepartmentsResponse struct {
	Departments []*Department `json:"departments"`
	TotalCount  int64         `json:"total_Count"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
}

func (d *Department) ToProto() *departmentpb.Department {
	dept := &departmentpb.Department{
		Id:            d.ID,
		Name:          d.Name,
		Description:   d.Description,
		Budget:        d.Budget,
		EmployeeCount: int32(d.EmployeeCount),
		CreatedAt:     timestamppb.New(d.CreatedAt),
		UpdatedAt:     timestamppb.New(d.UpdatedAt),
	}

	if d.ManagerID != nil {
		dept.ManagerId = *d.ManagerID
		if d.Manager != nil {
			dept.ManagerName = d.Manager.FirstName + " " + d.Manager.LastName
		}
	}

	return dept
}

func FromCreateRequest(req *CreateDepartmentRequest) *Department {
	return &Department{
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerID,
		Budget:      req.Budget,
		Location:    req.Location,
	}
}

func (d *Department) ApplyUpdate(req *UpdateDepartmentRequest) {
	if req.Name != "" {
		d.Name = req.Name
	}
	if req.Location != "" {
		d.Name = req.Name
	}
	if req.Description != "" {
		d.Description = req.Description
	}
	if req.ManagerID != nil {
		d.ManagerID = req.ManagerID
	}
	if req.Budget > 0 {
		d.Budget = req.Budget
	}
}

func (d *Department) HasManager() bool {
	return d.ManagerID != nil && *d.ManagerID != ""
}

func (d *Department) GetManagerName() string {
	if d.Manager != nil {
		return d.Manager.FirstName + " " + d.Manager.LastName
	}
	return ""
}

// IsActive returns true if the department is active (not soft deleted)
func (d *Department) IsActive() bool {
	return !d.DeletedAt.Valid
}

// CanDelete returns true if the department can be safely deleted
func (d *Department) CanDelete() bool {
	return d.EmployeeCount == 0
}
