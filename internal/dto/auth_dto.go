package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserDetail struct {
	UserID             int     `json:"user_id"`
	Username           string  `json:"username"`
	UserType           string  `json:"user_type"`
	EmployeeNPK        *string `json:"employee_npk,omitempty"`
	EmployeeName       *string `json:"employee_name,omitempty"`
	EmployeePosition   *string `json:"employee_position,omitempty"`
	EmployeeDepartment *string `json:"employee_department,omitempty"`
	EmployeeArea       *string `json:"employee_area,omitempty"`
	Permissions        []string `json:"permissions"`
}

type LoginResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	User         UserDetail `json:"user"`
}