package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---------- 服务项目 ----------

type serviceIn struct {
	Name        string `json:"name" binding:"required"`
	DurationMin int    `json:"duration_min" binding:"required,min=5"`
	PriceCents  int64  `json:"price_cents" binding:"required,min=0"`
	Active      *bool  `json:"active"`
}

func listServices(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []ServiceItem
		db.Order("id").Find(&items)
		c.JSON(http.StatusOK, items)
	}
}

func createService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in serviceIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item := ServiceItem{Name: in.Name, DurationMin: in.DurationMin, PriceCents: in.PriceCents, Active: true}
		if in.Active != nil {
			item.Active = *in.Active
		}
		db.Create(&item)
		c.JSON(http.StatusOK, item)
	}
}

func updateService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var item ServiceItem
		if err := db.First(&item, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "服务不存在"})
			return
		}
		var in serviceIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		item.Name, item.DurationMin, item.PriceCents = in.Name, in.DurationMin, in.PriceCents
		if in.Active != nil {
			item.Active = *in.Active
		}
		db.Save(&item)
		c.JSON(http.StatusOK, item)
	}
}

func deleteService(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 有历史预约的服务不物理删除，只停用，保证历史记录可追溯
		var n int64
		db.Model(&Booking{}).Where("service_item_id = ?", c.Param("id")).Count(&n)
		if n > 0 {
			db.Model(&ServiceItem{}).Where("id = ?", c.Param("id")).Update("active", false)
			c.JSON(http.StatusOK, gin.H{"ok": true, "deactivated": true})
			return
		}
		db.Delete(&ServiceItem{}, c.Param("id"))
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// ---------- 技师 ----------

type staffIn struct {
	Name   string `json:"name" binding:"required"`
	Phone  string `json:"phone"`
	Title  string `json:"title"`
	Active *bool  `json:"active"`
}

func listStaff(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []Staff
		db.Order("id").Find(&items)
		c.JSON(http.StatusOK, items)
	}
}

func createStaff(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in staffIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s := Staff{Name: in.Name, Phone: in.Phone, Title: in.Title, Active: true}
		if in.Active != nil {
			s.Active = *in.Active
		}
		db.Create(&s)
		c.JSON(http.StatusOK, s)
	}
}

func updateStaff(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s Staff
		if err := db.First(&s, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "技师不存在"})
			return
		}
		var in staffIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		s.Name, s.Phone, s.Title = in.Name, in.Phone, in.Title
		if in.Active != nil {
			s.Active = *in.Active
		}
		db.Save(&s)
		c.JSON(http.StatusOK, s)
	}
}

func deleteStaff(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var n int64
		db.Model(&Booking{}).Where("staff_id = ?", c.Param("id")).Count(&n)
		if n > 0 {
			db.Model(&Staff{}).Where("id = ?", c.Param("id")).Update("active", false)
			c.JSON(http.StatusOK, gin.H{"ok": true, "deactivated": true})
			return
		}
		db.Delete(&Staff{}, c.Param("id"))
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// ---------- 会员 ----------

type memberIn struct {
	Name            string `json:"name" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	DiscountPercent int    `json:"discount_percent"`
}

func listMembers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []Member
		db.Order("id desc").Find(&items)
		c.JSON(http.StatusOK, items)
	}
}

func createMember(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in memberIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if in.DiscountPercent <= 0 || in.DiscountPercent > 100 {
			in.DiscountPercent = 100
		}
		var n int64
		db.Model(&Member{}).Where("phone = ?", in.Phone).Count(&n)
		if n > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该手机号已是会员"})
			return
		}
		m := Member{Name: in.Name, Phone: in.Phone, DiscountPerson: in.DiscountPercent, CreatedAt: time.Now().Unix()}
		db.Create(&m)
		c.JSON(http.StatusOK, m)
	}
}

type rechargeIn struct {
	AmountCents int64  `json:"amount_cents" binding:"required,min=1"`
	Note        string `json:"note"`
}

// recharge 会员充值。当前阶段支付网关为模拟实现（直接记账），
// 对接微信/支付宝时只需把「记账前」的一段替换为真实支付回调。
func recharge(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var m Member
		if err := db.First(&m, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "会员不存在"})
			return
		}
		var in rechargeIn
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var txn Transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			m.BalanceCents += in.AmountCents
			if err := tx.Model(&Member{}).Where("id = ?", m.ID).Update("balance_cents", m.BalanceCents).Error; err != nil {
				return err
			}
			txn = Transaction{
				MemberID: m.ID, Type: TxnRecharge, AmountCents: in.AmountCents,
				BalanceAfter: m.BalanceCents, CreatedAt: time.Now().Unix(),
				Note: in.Note,
			}
			if in.Note == "" {
				txn.Note = "模拟支付网关收款"
			}
			return tx.Create(&txn).Error
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "记账失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"member": m, "transaction": txn})
	}
}

func memberTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []Transaction
		db.Where("member_id = ?", c.Param("id")).Order("id desc").Limit(100).Find(&items)
		c.JSON(http.StatusOK, items)
	}
}

func listTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		var items []Transaction
		db.Order("id desc").Limit(limit).Find(&items)
		c.JSON(http.StatusOK, items)
	}
}
