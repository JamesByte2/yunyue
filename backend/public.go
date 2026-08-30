package main

// 客户端公开接口：小程序 / H5 顾客自助预约使用，无需登录。
// 设计要点：只暴露最小必要信息（不含会员、流水等经营数据）；
// 写操作走与管理端同一套冲突检测与状态机，杜绝绕过。

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type publicBookingIn struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required,len=11"`
	ServiceItemID uint   `json:"service_item_id" binding:"required"`
	StaffID       uint   `json:"staff_id" binding:"required"`
	BookDate      string `json:"book_date" binding:"required"`
	StartMin      int    `json:"start_min" binding:"required,min=0,max=1439"`
}

type slot struct {
	StartMin  int  `json:"start_min"`
	Available bool `json:"available"`
}

func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

func timeNow() int64 { return time.Now().Unix() }

// PublicRoutes 顾客端开放能力。
func PublicRoutes(r *gin.Engine) {
	pub := r.Group("/api/public")
	{
		pub.GET("/catalog", publicCatalog(db))
		pub.GET("/availability", publicAvailability(db))
		pub.POST("/bookings", publicCreateBooking(db))
		pub.GET("/mybookings", publicMyBookings(db))
		pub.POST("/mybookings/:id/cancel", publicCancelBooking(db))
	}
}

// catalog：上架的服务 + 在岗技师。
func publicCatalog(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var services []ServiceItem
		db.Where("active = ?", true).Order("id").Find(&services)
		var staff []Staff
		db.Where("active = ?", true).Order("id").Find(&staff)
		c.JSON(http.StatusOK, gin.H{"services": services, "staff": staff})
	}
}

// availability：某技师某天 09:00-21:00 的 30 分钟粒度时段，与已有有效预约重叠的标记不可约。
func publicAvailability(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		date := c.Query("date")
		staffID := c.Query("staff_id")
		svcID := parseUint(c.Query("service_id"))
		if !validateDate(date) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式应为 YYYY-MM-DD"})
			return
		}
		sid := parseUint(staffID)
		var svc ServiceItem
		if err := db.First(&svc, svcID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "服务不存在"})
			return
		}
		duration := svc.DurationMin

		var actives []Booking
		db.Where("staff_id = ? AND book_date = ? AND status IN ?", sid, date,
			[]string{BookingPending, BookingConfirmed}).Find(&actives)

		const openMin, closeMin, step = 9 * 60, 21 * 60, 30
		slots := make([]slot, 0, (closeMin-openMin)/step)
		for m := openMin; m+duration <= closeMin; m += step {
			available := true
			for _, b := range actives {
				var d int
				var s ServiceItem
				db.Select("duration_min").First(&s, b.ServiceItemID)
				d = s.DurationMin
				if m < b.StartMin+d && b.StartMin < m+duration {
					available = false
					break
				}
			}
			slots = append(slots, slot{StartMin: m, Available: available})
		}
		c.JSON(http.StatusOK, gin.H{"date": date, "duration_min": duration, "slots": slots})
	}
}

func publicCreateBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in publicBookingIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !validateDate(in.BookDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式应为 YYYY-MM-DD"})
			return
		}
		// 复用管理端的冲突检测逻辑
		var svc ServiceItem
		if err := db.First(&svc, in.ServiceItemID).Error; err != nil || !svc.Active {
			c.JSON(http.StatusBadRequest, gin.H{"error": "服务不存在或已停用"})
			return
		}
		var staff Staff
		if err := db.First(&staff, in.StaffID).Error; err != nil || !staff.Active {
			c.JSON(http.StatusBadRequest, gin.H{"error": "技师不存在或已停用"})
			return
		}
		endMin := in.StartMin + svc.DurationMin
		var actives []Booking
		db.Where("staff_id = ? AND book_date = ? AND status IN ?",
			in.StaffID, in.BookDate, []string{BookingPending, BookingConfirmed}).Find(&actives)
		for _, b := range actives {
			var s ServiceItem
			db.Select("duration_min").First(&s, b.ServiceItemID)
			if in.StartMin < b.StartMin+s.DurationMin && b.StartMin < endMin {
				c.JSON(http.StatusConflict, gin.H{"error": "该时段刚被约满，请选择其他时间"})
				return
			}
		}
		b := Booking{
			CustomerName: in.CustomerName, CustomerPhone: in.CustomerPhone,
			ServiceItemID: in.ServiceItemID, StaffID: in.StaffID,
			BookDate: in.BookDate, StartMin: in.StartMin,
			Status: BookingPending, Remark: "客户自助预约",
			CreatedAt: timeNow(),
		}
		db.Create(&b)
		c.JSON(http.StatusOK, gin.H{"id": b.ID, "status": b.Status})
	}
}

// myBookings：凭手机号查自己的预约（手机号即轻量凭证）。
func publicMyBookings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		phone := c.Query("phone")
		if len(phone) != 11 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 11 位手机号"})
			return
		}
		var list []Booking
		db.Where("customer_phone = ?", phone).Order("id desc").Limit(20).Find(&list)
		var services []ServiceItem
		db.Find(&services)
		name := map[uint]string{}
		for _, s := range services {
			name[s.ID] = s.Name
		}
		out := make([]gin.H, 0, len(list))
		for _, b := range list {
			out = append(out, gin.H{
				"id": b.ID, "book_date": b.BookDate, "start_min": b.StartMin,
				"service_name": name[b.ServiceItemID], "status": b.Status,
			})
		}
		sort.Slice(out, func(i, j int) bool { return out[i]["id"].(uint) > out[j]["id"].(uint) })
		c.JSON(http.StatusOK, out)
	}
}

// cancel：仅本人手机号 + 仅待确认状态可取消。
func publicCancelBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in struct {
			Phone string `json:"phone" binding:"required,len=11"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var b Booking
		if err := db.First(&b, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		if b.CustomerPhone != in.Phone {
			c.JSON(http.StatusForbidden, gin.H{"error": "手机号不匹配，无法取消"})
			return
		}
		if b.Status != BookingPending {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅待确认的预约可以取消，请联系门店"})
			return
		}
		db.Model(&b).Update("status", BookingCanceled)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
