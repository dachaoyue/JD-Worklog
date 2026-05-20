package handlers

import (
	"net/http"
	"time"

	"github.com/example/worklog-system/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkScheduleHandler struct{ DB *gorm.DB }

type scheduleDayResp struct {
	Date      string `json:"date"`
	IsHoliday bool   `json:"is_holiday"`
	Weekday   int    `json:"weekday"` // 0=Sunday
}

type scheduleTaskResp struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	PersonName    string `json:"person_name"`
	TechDirection string `json:"tech_direction"`
	Color         string `json:"color"`
	Label         string `json:"label"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	WorkDays      int    `json:"work_days"`
}

func normalizeColor(c string) string {
	if c == "" {
		return "#409eff"
	}
	return c
}

func parseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func defaultIsHoliday(d time.Time) bool {
	wd := d.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}

func (h *WorkScheduleHandler) loadHolidayMap(start, end time.Time) map[string]bool {
	var rows []models.WorkScheduleDay
	h.DB.Where("date >= ? AND date <= ?", start, end).Find(&rows)
	m := make(map[string]bool, len(rows))
	for _, r := range rows {
		m[formatDate(r.Date)] = r.IsHoliday
	}
	return m
}

func effectiveHoliday(d time.Time, overrides map[string]bool) bool {
	key := formatDate(d)
	if v, ok := overrides[key]; ok {
		return v
	}
	return defaultIsHoliday(d)
}

func countWorkDays(start, end time.Time, overrides map[string]bool) int {
	n := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if !effectiveHoliday(d, overrides) {
			n++
		}
	}
	return n
}

type scheduleUserResp struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// GET /api/work-schedules/users
func (h *WorkScheduleHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Order("id asc").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db"})
		return
	}
	resp := make([]scheduleUserResp, 0, len(users))
	for _, u := range users {
		resp = append(resp, scheduleUserResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// GET /api/work-schedules/board?start=&end=
func (h *WorkScheduleHandler) GetBoard(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end required"})
		return
	}
	start, err := parseDate(startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}
	end, err := parseDate(endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}
	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be >= start"})
		return
	}
	daysBetween := int(end.Sub(start).Hours()/24) + 1
	if daysBetween > 93 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date range must not exceed 93 days"})
		return
	}

	overrides := h.loadHolidayMap(start, end)

	var days []scheduleDayResp
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, scheduleDayResp{
			Date:      formatDate(d),
			IsHoliday: effectiveHoliday(d, overrides),
			Weekday:   int(d.Weekday()),
		})
	}

	var tasks []models.WorkScheduleTask
	h.DB.Preload("User").Where("end_date >= ? AND start_date <= ?", start, end).
		Order("user_id asc, start_date asc, id asc").Find(&tasks)

	taskResps := make([]scheduleTaskResp, 0, len(tasks))
	for _, t := range tasks {
		name := t.User.Nickname
		if name == "" {
			name = t.User.Username
		}
		taskResps = append(taskResps, scheduleTaskResp{
			ID:            t.ID,
			UserID:        t.UserID,
			PersonName:    name,
			TechDirection: t.TechDirection,
			Color:         normalizeColor(t.Color),
			Label:         t.Label,
			StartDate:     formatDate(t.StartDate),
			EndDate:       formatDate(t.EndDate),
			WorkDays:      countWorkDays(t.StartDate, t.EndDate, overrides),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"start":       startStr,
		"end":         endStr,
		"today":       formatDate(time.Now()),
		"days":        days,
		"tasks":       taskResps,
		"task_count":  len(taskResps),
		"project_count": countDistinctUsers(tasks),
	})
}

func countDistinctUsers(tasks []models.WorkScheduleTask) int {
	seen := map[uint]bool{}
	for _, t := range tasks {
		seen[t.UserID] = true
	}
	return len(seen)
}

// 合并同一人员、同一项目（名称+颜色+说明）且日期相连/重叠的排期为一条
func (h *WorkScheduleHandler) mergeAdjacentSameProject(userID uint, color, techDirection, label string) {
	var tasks []models.WorkScheduleTask
	h.DB.Where("user_id = ? AND color = ? AND tech_direction = ? AND label = ?",
		userID, color, techDirection, label).Order("start_date asc").Find(&tasks)
	if len(tasks) < 2 {
		return
	}
	base := tasks[0]
	for i := 1; i < len(tasks); i++ {
		cur := tasks[i]
		nextDay := base.EndDate.AddDate(0, 0, 1)
		if !cur.StartDate.After(nextDay) {
			if cur.EndDate.After(base.EndDate) {
				base.EndDate = cur.EndDate
			}
			h.DB.Delete(&models.WorkScheduleTask{}, cur.ID)
		} else {
			h.DB.Save(&base)
			base = cur
		}
	}
	h.DB.Save(&base)
}

type createScheduleTaskReq struct {
	UserID        uint   `json:"user_id" binding:"required"`
	TechDirection string `json:"tech_direction" binding:"required"`
	Color         string `json:"color"`
	Label         string `json:"label"`
	StartDate     string `json:"start_date" binding:"required"`
	EndDate       string `json:"end_date" binding:"required"`
}

// POST /api/admin/work-schedules/tasks
func (h *WorkScheduleHandler) CreateTask(c *gin.Context) {
	var req createScheduleTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	if req.TechDirection == "" {
		req.TechDirection = "排期"
	}
	start, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}
	end, err := parseDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}
	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be >= start"})
		return
	}
	var user models.User
	if err := h.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
		return
	}
	task := models.WorkScheduleTask{
		UserID:        req.UserID,
		TechDirection: req.TechDirection,
		Color:         normalizeColor(req.Color),
		Label:         req.Label,
		StartDate:     start,
		EndDate:       end,
	}
	if err := h.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": task.ID})
}

type updateScheduleTaskReq struct {
	UserID        uint    `json:"user_id"`
	TechDirection string  `json:"tech_direction"`
	Color         string  `json:"color"`
	Label         *string `json:"label"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
}

// PUT /api/admin/work-schedules/tasks/:id
func (h *WorkScheduleHandler) UpdateTask(c *gin.Context) {
	var task models.WorkScheduleTask
	if err := h.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var req updateScheduleTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	if req.UserID != 0 {
		task.UserID = req.UserID
	}
	if req.TechDirection != "" {
		task.TechDirection = req.TechDirection
	}
	if req.Color != "" {
		task.Color = normalizeColor(req.Color)
	}
	if req.Label != nil {
		task.Label = *req.Label
	}
	if req.StartDate != "" {
		start, err := parseDate(req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
			return
		}
		task.StartDate = start
	}
	if req.EndDate != "" {
		end, err := parseDate(req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
			return
		}
		task.EndDate = end
	}
	if task.EndDate.Before(task.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end must be >= start"})
		return
	}
	if err := h.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/admin/work-schedules/tasks/:id
func (h *WorkScheduleHandler) DeleteTask(c *gin.Context) {
	if err := h.DB.Delete(&models.WorkScheduleTask{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type toggleCalendarReq struct {
	Date string `json:"date" binding:"required"`
}

// POST /api/admin/work-schedules/calendar/toggle
func (h *WorkScheduleHandler) ToggleCalendarDay(c *gin.Context) {
	var req toggleCalendarReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
		return
	}
	d, err := parseDate(req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	overrides := h.loadHolidayMap(d, d)
	current := effectiveHoliday(d, overrides)
	next := !current

	var row models.WorkScheduleDay
	err = h.DB.Where("date = ?", d).First(&row).Error
	if err == gorm.ErrRecordNotFound {
		row = models.WorkScheduleDay{Date: d, IsHoliday: next}
		h.DB.Create(&row)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		row.IsHoliday = next
		h.DB.Save(&row)
	}
	c.JSON(http.StatusOK, gin.H{"date": req.Date, "is_holiday": next})
}
