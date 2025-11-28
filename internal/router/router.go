package router

import (
	"e-memo-job-reservation-api/internal/auth"
	"e-memo-job-reservation-api/internal/handler"
	"e-memo-job-reservation-api/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type AllHandlers struct {
	AuthHandler                *handler.AuthHandler
	DepartmentHandler          *handler.DepartmentHandler
	AreaHandler                *handler.AreaHandler
	EmployeeHandler            *handler.EmployeeHandler
	PhysicalLocationHandler    *handler.PhysicalLocationHandler
	AccessPermissionHandler    *handler.AccessPermissionHandler
	SectionStatusTicketHandler *handler.SectionStatusTicketHandler
	StatusTicketHandler        *handler.StatusTicketHandler
	TicketHandler              *handler.TicketHandler
	PositionPermissionHandler  *handler.PositionPermissionHandler
	EmployeePositionHandler    *handler.EmployeePositionHandler
	WorkflowHandler            *handler.WorkflowHandler
	SpecifiedLocationHandler   *handler.SpecifiedLocationHandler
	RejectedTicketHandler      *handler.RejectedTicketHandler
	JobHandler                 *handler.JobHandler
	ActionHandler              *handler.ActionHandler
	FileHandler                *handler.FileHandler
	SystemHandler              *handler.SystemHandler
}

type AllRepositories struct {
	PositionPermissionRepo *repository.PositionPermissionRepository
}

func SetupRouter(h *AllHandlers, r *AllRepositories, authMiddleware *auth.AuthMiddleware, wsHandler *handler.WebSocketHandler, editModeMiddleware *auth.EditModeMiddleware) *gin.Engine {
	router := gin.Default()

	// Allow all origins - no security restrictions
	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Access-Control-Request-Method", "Access-Control-Request-Headers"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false, // Must be false when AllowAllOrigins is true
		MaxAge:           12 * 3600,
	}
	router.Use(cors.New(config))

	api := router.Group("/api/e-memo-job-reservation")

	public := api.Group("")
	{
		public.POST("/login", h.AuthHandler.Login)
		public.POST("/refresh", h.AuthHandler.RefreshToken)

		// DEPARTMENT
		public.GET("/departments", h.DepartmentHandler.GetAllDepartments)
		public.GET("/department/:id", h.DepartmentHandler.GetDepartmentByID)
		public.GET("/department/options", h.DepartmentHandler.GetRequestorDepartmentOptions)

		// AREA
		public.GET("/areas", h.AreaHandler.GetAllAreas)
		public.GET("/area/:id", h.AreaHandler.GetAreaByID)

		// PHYSICAL LOCATION
		public.GET("/physical-location", h.PhysicalLocationHandler.GetAllPhysicalLocations)
		public.GET("/physical-location/:id", h.PhysicalLocationHandler.GetPhysicalLocationByID)

		// SPECIFIED LOCATION
		public.GET("/specified-location", h.SpecifiedLocationHandler.GetAllSpecifiedLocations)
		public.GET("/specified-location/:id", h.SpecifiedLocationHandler.GetSpecifiedLocationByID)

		// EMPLOYEE
		public.GET("/employee", h.EmployeeHandler.GetAllEmployees)
		public.GET("/employee/options", h.EmployeeHandler.GetEmployeeOptions)

		// EMPLOYEE POSITION
		public.GET("/employee-positions", h.EmployeePositionHandler.GetAllEmployeePositions)
		public.GET("/employee-positions/:id", h.EmployeePositionHandler.GetEmployeePositionByID)

		// TICKET
		public.GET("/tickets", h.TicketHandler.GetAllTickets)
		public.GET("/tickets/:id", h.TicketHandler.GetTicketByID)

		// STATUS TICKET
		public.GET("/status-ticket", h.StatusTicketHandler.GetAllStatusTickets)

		// ACTIONS
		public.GET("/actions", h.ActionHandler.GetAllActions)

		// WEB SOCKET
		public.GET("/ws", wsHandler.ServeWs)
		public.POST("/auth/ws-public-ticket", h.AuthHandler.GeneratePublicWebSocketTicket)
	}

	reportRoutes := api.Group("/reports")
	{
		reportRoutes.GET("/ticket-summary", h.TicketHandler.GetTicketSummary)
		reportRoutes.GET("/oldest-ticket", h.TicketHandler.GetOldestTicket)
	}

	private := api.Group("")
	private.Use(authMiddleware.JWTMiddleware())
	{
		private.POST("/logout", h.AuthHandler.Logout)
		private.GET("/files/download", h.FileHandler.DownloadFile)
		private.GET("/files/view", h.FileHandler.ViewFile)
		private.POST("/auth/ws-ticket", h.AuthHandler.GenerateWebSocketTicket)

		systemRoutes := private.Group("/system")
		systemRoutes.Use(auth.RequirePermission("MASTER_USER", r.PositionPermissionRepo))
		{
			systemRoutes.POST("/edit-mode", h.SystemHandler.UpdateEditMode)
		}

		setupMasterDataRoutes(private, h, r)

		setupMainDataRoutes(private, h, r, editModeMiddleware)
	}

	return router
}

func setupMasterDataRoutes(group *gin.RouterGroup, h *AllHandlers, r *AllRepositories) {
	masterGroup := group.Group("")
	masterGroup.Use(auth.RequirePermission("MASTER_USER", r.PositionPermissionRepo))
	{
		deptRoutes := masterGroup.Group("/department")
		{
			deptRoutes.POST("", h.DepartmentHandler.CreateDepartment)
			deptRoutes.DELETE("/:id", h.DepartmentHandler.DeleteDepartment)
			deptRoutes.PUT("/:id", h.DepartmentHandler.UpdateDepartment)
			deptRoutes.PATCH("/:id/status", h.DepartmentHandler.UpdateDepartmentActiveStatus)
		}

		areaRoutes := masterGroup.Group("/area")
		{
			areaRoutes.POST("", h.AreaHandler.CreateArea)
			areaRoutes.PUT("/:id", h.AreaHandler.UpdateArea)
			areaRoutes.PATCH("/:id/status", h.AreaHandler.UpdateAreaActiveStatus)
		}

		employeeRoutes := masterGroup.Group("/employee")
		{
			employeeRoutes.POST("", h.EmployeeHandler.CreateEmployee)
			employeeRoutes.POST("/batch", h.EmployeeHandler.BatchProcessEmployees)
			employeeRoutes.PUT("/:npk", h.EmployeeHandler.UpdateEmployee)
			employeeRoutes.PATCH("/:npk/status", h.EmployeeHandler.UpdateEmployeeActiveStatus)
		}

		physicalLocationRoutes := group.Group("/physical-location")
		{
			physicalLocationRoutes.POST("", h.PhysicalLocationHandler.CreatePhysicalLocation)
			physicalLocationRoutes.PUT("/:id", h.PhysicalLocationHandler.UpdatePhysicalLocation)
			physicalLocationRoutes.DELETE("/:id", h.PhysicalLocationHandler.DeletePhysicalLocation)
			physicalLocationRoutes.PATCH("/:id/status", h.PhysicalLocationHandler.UpdatePhysicalLocationActiveStatus)
		}

		specLocRoutes := group.Group("/specified-location")
		{
			specLocRoutes.POST("", h.SpecifiedLocationHandler.CreateSpecifiedLocation)
			specLocRoutes.PUT("/:id", h.SpecifiedLocationHandler.UpdateSpecifiedLocation)
			specLocRoutes.DELETE("/:id", h.SpecifiedLocationHandler.DeleteSpecifiedLocation)
			specLocRoutes.PATCH("/:id/status", h.SpecifiedLocationHandler.UpdateSpecifiedLocationActiveStatus)
		}

		accessPermissionRoutes := group.Group("/access-permission")
		{
			accessPermissionRoutes.POST("", h.AccessPermissionHandler.CreateAccessPermission)
			accessPermissionRoutes.GET("", h.AccessPermissionHandler.GetAllAccessPermissions)
			accessPermissionRoutes.GET("/:id", h.AccessPermissionHandler.GetAccessPermissionByID)
			accessPermissionRoutes.PUT("/:id", h.AccessPermissionHandler.UpdateAccessPermission)
			accessPermissionRoutes.DELETE("/:id", h.AccessPermissionHandler.DeleteAccessPermission)
			accessPermissionRoutes.PATCH("/:id/status", h.AccessPermissionHandler.UpdateAccessPermissionActiveStatus)
		}

		statusTicketRoutes := group.Group("/status-ticket")
		{
			statusTicketRoutes.POST("", h.StatusTicketHandler.CreateStatusTicket)
			statusTicketRoutes.GET("/:id", h.StatusTicketHandler.GetStatusTicketByID)
			statusTicketRoutes.DELETE("/:id", h.StatusTicketHandler.DeleteStatusTicket)
			statusTicketRoutes.PATCH("/:id/status", h.StatusTicketHandler.UpdateStatusTicketActiveStatus)
			statusTicketRoutes.PUT("/reorder", h.StatusTicketHandler.ReorderStatusTickets)
		}

		sectionRoutes := group.Group("/section-status-ticket")
		{
			sectionRoutes.POST("", h.SectionStatusTicketHandler.CreateSectionStatusTicket)
			sectionRoutes.GET("", h.SectionStatusTicketHandler.GetAllSectionStatusTickets)
			sectionRoutes.GET("/:id", h.SectionStatusTicketHandler.GetSectionStatusTicketByID)
			sectionRoutes.PATCH("/:id/status", h.SectionStatusTicketHandler.UpdateSectionStatusTicketActiveStatus)
			sectionRoutes.PUT("/:id", h.SectionStatusTicketHandler.UpdateSectionStatusTicket)
			sectionRoutes.DELETE("/:id", h.SectionStatusTicketHandler.DeleteSectionStatusTicket)
			sectionRoutes.PUT("/reorder", h.SectionStatusTicketHandler.ReorderSections)
		}

		posPermRoutes := group.Group("/position-permissions")
		{
			posPermRoutes.POST("", h.PositionPermissionHandler.CreatePositionPermission)
			posPermRoutes.GET("", h.PositionPermissionHandler.GetAllPositionPermissions)
			posPermRoutes.PATCH("/positions/:posId/permissions/:permId/status", h.PositionPermissionHandler.UpdatePositionPermissionActiveStatus)
			posPermRoutes.DELETE("/positions/:posId/permissions/:permId", h.PositionPermissionHandler.DeletePositionPermission)
		}
		posRoutes := group.Group("/employee-position")
		{
			posRoutes.POST("", h.EmployeePositionHandler.CreateEmployeePosition)
			posRoutes.PUT("/:id", h.EmployeePositionHandler.UpdateEmployeePosition)
			posRoutes.DELETE("/:id", h.EmployeePositionHandler.DeleteEmployeePosition)
			posRoutes.PATCH("/:id/status", h.EmployeePositionHandler.UpdateEmployeePositionActiveStatus)
		}
		workflowRoutes := group.Group("/workflow")
		{
			workflowRoutes.POST("", h.WorkflowHandler.CreateWorkflow)
			workflowRoutes.GET("", h.WorkflowHandler.GetAllWorkflows)
			workflowRoutes.GET("/:id", h.WorkflowHandler.GetWorkflowByID)
			workflowRoutes.PUT("/:id", h.WorkflowHandler.UpdateWorkflow)
			workflowRoutes.DELETE("/:id", h.WorkflowHandler.DeleteWorkflow)
			workflowRoutes.PATCH("/:id/status", h.WorkflowHandler.UpdateWorkflowActiveStatus)

			stepRoutes := group.Group("/workflow-step")
			{
				stepRoutes.POST("", h.WorkflowHandler.AddWorkflowStep)
				stepRoutes.GET("", h.WorkflowHandler.GetAllWorkflowSteps)
				stepRoutes.GET("/:id", h.WorkflowHandler.GetWorkflowStepByID)
				stepRoutes.DELETE("/:id", h.WorkflowHandler.DeleteWorkflowStep)
				stepRoutes.PATCH("/:id/status", h.WorkflowHandler.UpdateWorkflowStepActiveStatus)
			}
		}
	}
}

// MAIN TICKET
func setupMainDataRoutes(group *gin.RouterGroup, h *AllHandlers, r *AllRepositories, editModeMiddleware *auth.EditModeMiddleware) {
	ticketRoutes := group.Group("/tickets")
	{
		ticketRoutes.POST("", editModeMiddleware.CheckEditMode(), auth.RequirePermission("CREATE_TICKET", r.PositionPermissionRepo), h.TicketHandler.CreateTicket)
		ticketRoutes.PUT("/:id", editModeMiddleware.CheckEditMode(), h.TicketHandler.UpdateTicket)
		ticketRoutes.PUT("/reorder", editModeMiddleware.CheckEditMode(), auth.RequirePermission("TICKET_PRIORITY_MANAGE", r.PositionPermissionRepo), h.TicketHandler.ReorderTickets)
		ticketRoutes.POST("/:id/action", editModeMiddleware.CheckEditMode(), h.TicketHandler.ExecuteAction)
		ticketRoutes.GET("/:id/available-actions", h.TicketHandler.GetAvailableActions)
		ticketRoutes.GET("/:id/files", h.FileHandler.GetAllFilesByTicketID)
		ticketRoutes.POST("/:id/files", editModeMiddleware.CheckEditMode(), h.TicketHandler.AddSupportFiles)
		ticketRoutes.DELETE("/:id/files", editModeMiddleware.CheckEditMode(), h.TicketHandler.RemoveSupportFiles)
		ticketRoutes.GET("/:id/last-rejection", h.TicketHandler.GetLastRejectionDetail)
	}

	jobRoutes := group.Group("/jobs")
	{
		jobRoutes.GET("", h.JobHandler.GetAllJobs)
		jobRoutes.GET("/:id", h.JobHandler.GetJobByID)
		jobRoutes.GET("/:id/available-actions", h.JobHandler.GetAvailableActions)
		jobRoutes.PUT("/:id/assign", editModeMiddleware.CheckEditMode(), auth.RequirePermission("JOB_ASSIGN_PIC", r.PositionPermissionRepo), h.JobHandler.AssignPIC)
		jobRoutes.PUT("/reorder", editModeMiddleware.CheckEditMode(), auth.RequirePermission("JOB_PRIORITY_MANAGE", r.PositionPermissionRepo), h.JobHandler.ReorderJobs)
	}
}
