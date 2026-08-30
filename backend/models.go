package main

// 领域模型。金额一律用「分」存储，避免浮点误差；折扣用整数百分比（100 = 无折扣）。

type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Email        string `gorm:"size:128;uniqueIndex" json:"email"`
	PasswordHash string `gorm:"size:128" json:"-"`
	Role         string `gorm:"size:16" json:"role"` // admin | staff
	Name         string `gorm:"size:64" json:"name"`
}

type ServiceItem struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:64" json:"name"`
	DurationMin int    `json:"duration_min"`
	PriceCents  int64  `json:"price_cents"`
	Active      bool   `gorm:"default:true" json:"active"`
}

type Staff struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `gorm:"size:64" json:"name"`
	Phone  string `gorm:"size:32" json:"phone"`
	Title  string `gorm:"size:64" json:"title"`
	Active bool   `gorm:"default:true" json:"active"`
}

const (
	BookingPending   = "pending"
	BookingConfirmed = "confirmed"
	BookingDone      = "done"
	BookingCanceled  = "canceled"
)

type Booking struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	CustomerName  string `gorm:"size:64" json:"customer_name"`
	CustomerPhone string `gorm:"size:32" json:"customer_phone"`
	ServiceItemID uint   `gorm:"index" json:"service_item_id"`
	StaffID       uint   `gorm:"index" json:"staff_id"`
	BookDate      string `gorm:"size:10;index" json:"book_date"` // YYYY-MM-DD
	StartMin      int    `json:"start_min"`                      // 距当日 00:00 的分钟数
	Status        string `gorm:"size:16;index" json:"status"`
	Remark        string `gorm:"size:255" json:"remark"`
	MemberID      *uint  `json:"member_id"`
	PayableCents  int64  `json:"payable_cents"` // 成交金额（完成时计算）
	CreatedAt     int64  `json:"created_at"`
}

type Member struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	Name           string `gorm:"size:64" json:"name"`
	Phone          string `gorm:"size:32;uniqueIndex" json:"phone"`
	BalanceCents   int64  `json:"balance_cents"`
	DiscountPerson int    `gorm:"default:100" json:"discount_percent"` // 90 = 九折
	CreatedAt      int64  `json:"created_at"`
}

const (
	TxnRecharge = "recharge"
	TxnConsume  = "consume"
)

type Transaction struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	MemberID     uint   `gorm:"index" json:"member_id"`
	BookingID    *uint  `json:"booking_id"`
	Type         string `gorm:"size:16" json:"type"`
	AmountCents  int64  `json:"amount_cents"`
	BalanceAfter int64  `json:"balance_after"`
	Note         string `gorm:"size:255" json:"note"`
	CreatedAt    int64  `json:"created_at"`
}
