package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kidkender/prohealth-api/internal/handler"
	"github.com/kidkender/prohealth-api/internal/middleware"
	"github.com/kidkender/prohealth-api/internal/model"
)

func RegisterScheduleRoutes(router gin.IRouter, scheduleHandler *handler.ScheduleHandler) {
	schedule := router.Group("/schedule")
	schedule.GET("/period", scheduleHandler.GetSchedulePeriods)
	schedule.GET("/period/:id", scheduleHandler.GetSchedulePeriodById)
	schedule.POST("/period",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.CreatePeriod,
	)
	schedule.PUT("/period/:id",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.UpdatePeriod,
	)
	schedule.DELETE("/period/:id",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.DeleteSchedulePeriod,
	)
	schedule.GET("/shift", scheduleHandler.GetAllShifts)
	schedule.GET("/shift/:id", scheduleHandler.GetScheduleShiftByID)
	schedule.GET("/period/:id/shift", scheduleHandler.GetShiftsByPeriod)
	schedule.POST("/shift",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.CreateShift,
	)
	schedule.GET("/shift/doctor/:id", scheduleHandler.GetScheduleShiftByDoctor)
	schedule.PUT("/shift/doctor",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.SetDoctorShift,
	)

	schedule.PUT("/shift/:id",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.UpdateScheduleShift,
	)
	schedule.PUT("/shift/:id/doctors",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.ReplaceDoctorsForShift,
	)
	schedule.DELETE("/shift/:id",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.DeleteScheduleShift,
	)

	schedule.POST("/generate",
		middleware.RequireRole(model.RoleAdmin),
		scheduleHandler.GenerateSchedule,
	)

}
