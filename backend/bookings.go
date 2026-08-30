package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 可用预约状态集合 + 合法流转。终态（done/canceled）不可再变更。
var transitions = map[string][]string{
	BookingPending:   {BookingConfirmed, BookingCanceled},
	BookingConfirmed: {BookingDone, BookingCanceled},
}

type bookingIn struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	ServiceItemID uint   `json:"service_item_id" binding:"required"`
	StaffID       uint   `json:"staff_id" binding:"required"`
	BookDate      string `json:"book_date" binding:"required"` // YYYY-MM-DD
	StartMin      int    `json:"start_min" binding:"required,min=0,max=1439"`
	Remark        string `json:"remark"`
}

func validateDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

type bookingView struct {
	Booking
	ServiceName string `json:"service_name"`
	StaffName   string `json:"staff_name"`
}

// listBookings 返回带服务名/技师名的视图，支持 date/status 过滤。
func listBookings(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := db.Model(&Booking{})
		if d := c.Query("date"); d != "" {
			q = q.Where("book_date = ?", d)
		}
		if s := c.Query("status"); s != "" {
			q = q.Where("status = ?", s)
		}
		var bookings []Booking
		q.Order("book_date desc, start_min").Limit(200).Find(&bookings)

		var services []ServiceItem
		db.Find(&services)
		svcName := map[uint]string{}
		for _, s := range services {
			svcName[s.ID] = s.Name
		}
		var staffs []Staff
		db.Find(&staffs)
		staffName := map[uint]string{}
		for _, s := range staffs {
			staffName[s.ID] = s.Name
		}

		out := make([]bookingView, 0, len(bookings))
		for _, b := range bookings {
			out = append(out, bookingView{Booking: b, ServiceName: svcName[b.ServiceItemID], StaffName: staffName[b.StaffID]})
		}
		c.JSON(http.StatusOK, out)
	}
}

// createBooking 核心校验：服务/技师有效、日期合法、同技师同时段无有效预约重叠。
func createBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in bookingIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if !validateDate(in.BookDate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式应为 YYYY-MM-DD"})
			return
		}
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
		if endMin > 24*60 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "预约结束时间超出营业日"})
			return
		}

		// 冲突检测：同技师、同日期、未完结状态，时间段 [start, end) 有交集即冲突
		var actives []Booking
		db.Where(
			"staff_id = ? AND book_date = ? AND status IN ?",
			in.StaffID, in.BookDate, []string{BookingPending, BookingConfirmed},
		).Find(&actives)
		for _, b := range actives {
			var dur int
			db.Model(&ServiceItem{}).Select("duration_min").First(&svc, b.ServiceItemID)
			dur = svc.DurationMin
			if in.StartMin < b.StartMin+dur && b.StartMin < endMin {
				c.JSON(http.StatusConflict, gin.H{
					"error":      fmt.Sprintf("与已有预约 #%d 时间冲突（%02d:%02d 起约 %d 分钟）", b.ID, b.StartMin/60, b.StartMin%60, dur),
					"conflictId": b.ID,
				})
				return
			}
		}

		b := Booking{
			CustomerName: in.CustomerName, CustomerPhone: in.CustomerPhone,
			ServiceItemID: in.ServiceItemID, StaffID: in.StaffID,
			BookDate: in.BookDate, StartMin: in.StartMin,
			Status: BookingPending, Remark: in.Remark,
			CreatedAt: time.Now().Unix(),
		}
		db.Create(&b)
		c.JSON(http.StatusOK, b)
	}
}

// updateBookingStatus 状态流转：pending→confirmed/canceled；confirmed→done/canceled；终态锁定。
func updateBookingStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b Booking
		if err := db.First(&b, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		var in struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		allowed, ok := transitions[b.Status]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该预约已完结，状态不可变更"})
			return
		}
		valid := false
		for _, s := range allowed {
			if s == in.Status {
				valid = true
			}
		}
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "不允许 " + b.Status + " → " + in.Status})
			return
		}
		db.Model(&b).Update("status", in.Status)
		c.JSON(http.StatusOK, b)
	}
}

type finishIn struct {
	MemberID *uint `json:"member_id"` // 提供则从会员卡扣费
}

// finishBooking 完成预约并结算：现金直接完结；会员卡按折扣扣款并生成消费流水。
// 结算放在数据库事务里：扣款与流水要么同时成功，要么都不发生。
func finishBooking(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b Booking
		if err := db.First(&b, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "预约不存在"})
			return
		}
		if b.Status != BookingConfirmed {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅已确认的预约可以完成结算"})
			return
		}
		var in finishIn
		_ = c.ShouldBindJSON(&in)

		var svc ServiceItem
		db.First(&svc, b.ServiceItemID)
		b.PayableCents = svc.PriceCents

		if in.MemberID != nil {
			var m Member
			if err := db.First(&m, *in.MemberID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
				return
			}
			payable := svc.PriceCents * int64(m.DiscountPerson) / 100
			if m.BalanceCents < payable {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":        fmt.Sprintf("会员余额不足：应付 %d 分（%.1f 折），余额 %d 分", payable, float64(m.DiscountPerson)/10, m.BalanceCents),
					"payable":      payable,
					"balance":      m.BalanceCents,
				})
				return
			}
			err := db.Transaction(func(tx *gorm.DB) error {
				m.BalanceCents -= payable
				if err := tx.Model(&Member{}).Where("id = ?", m.ID).Updates(map[string]any{
					"balance_cents": m.BalanceCents,
				}).Error; err != nil {
					return err
				}
				b.PayableCents = payable
				b.MemberID = &m.ID
				b.Status = BookingDone
				if err := tx.Save(&b).Error; err != nil {
					return err
				}
				return tx.Create(&Transaction{
					MemberID: m.ID, BookingID: &b.ID, Type: TxnConsume,
					AmountCents: payable, BalanceAfter: m.BalanceCents,
					Note:    fmt.Sprintf("预约 #%d %s", b.ID, svc.Name),
					CreatedAt: time.Now().Unix(),
				}).Error
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "结算失败"})
				return
			}
			c.JSON(http.StatusOK, b)
			return
		}

		b.Status = BookingDone
		db.Save(&b)
		c.JSON(http.StatusOK, b)
	}
}

// ---------- 工作台看板 ----------

func dashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		today := time.Now().Format("2006-01-02")
		month := time.Now().Format("2006-01")

		var todayCount, pendingCount, memberCount int64
		db.Model(&Booking{}).Where("book_date = ? AND status <> ?", today, BookingCanceled).Count(&todayCount)
		db.Model(&Booking{}).Where("status = ?", BookingPending).Count(&pendingCount)
		db.Model(&Member{}).Count(&memberCount)

		var monthRevenue int64
		db.Model(&Transaction{}).
			Where("type = ? AND created_at >= ?", TxnConsume, monthStartUnix(month)).
			Select("COALESCE(SUM(amount_cents),0)").Scan(&monthRevenue)

		var todayList []Booking
		db.Where("book_date = ? AND status <> ?", today, BookingCanceled).
			Order("start_min").Find(&todayList)

		c.JSON(http.StatusOK, gin.H{
			"today_bookings":   todayCount,
			"pending_count":    pendingCount,
			"month_revenue":    monthRevenue,
			"member_count":     memberCount,
			"today_list":       todayList,
		})
	}
}

func monthStartUnix(month string) int64 {
	t, _ := time.ParseInLocation("2006-01", month, time.Local)
	return t.Unix()
}

var _ = strconv.Itoa // 保留引用
