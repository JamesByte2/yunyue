package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	cfg := LoadConfig()

	var err error
	for i := 0; i < 10; i++ { // 容器编排下 MySQL 可能晚于本服务就绪，重试等待
		if db, err = gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{}); err == nil {
			break
		}
		log.Printf("waiting for database: %v", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("database unavailable: ", err)
	}

	if err := db.AutoMigrate(&User{}, &ServiceItem{}, &Staff{}, &Booking{}, &Member{}, &Transaction{}); err != nil {
		log.Fatal("migrate: ", err)
	}
	seed()

	r := gin.Default()
	r.Use(CORS())

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := r.Group("/api")
	api.POST("/auth/login", login(cfg.JWTSecret))
	PublicRoutes(r)

	auth := api.Group("/", AuthRequired(cfg.JWTSecret))
	{
		auth.GET("/me", func(c *gin.Context) { c.JSON(http.StatusOK, currentUser(db, c)) })

		// 预约：管理员与技师都可操作
		auth.GET("/bookings", listBookings(db))
		auth.POST("/bookings", createBooking(db))
		auth.PATCH("/bookings/:id/status", updateBookingStatus(db))
		auth.POST("/bookings/:id/finish", finishBooking(db))
		auth.GET("/dashboard", dashboard(db))

		// 只读资源：全员可见
		auth.GET("/services", listServices(db))
		auth.GET("/staff", listStaff(db))

		// 管理员写操作
		admin := auth.Group("/", AdminRequired())
		{
			admin.POST("/services", createService(db))
			admin.PUT("/services/:id", updateService(db))
			admin.DELETE("/services/:id", deleteService(db))

			admin.POST("/staff", createStaff(db))
			admin.PUT("/staff/:id", updateStaff(db))
			admin.DELETE("/staff/:id", deleteStaff(db))

			admin.GET("/members", listMembers(db))
			admin.POST("/members", createMember(db))
			admin.POST("/members/:id/recharge", recharge(db))
			admin.GET("/members/:id/transactions", memberTransactions(db))
			admin.GET("/transactions", listTransactions(db))
		}
	}

	log.Println("yunyue-api listening on :" + cfg.Port)
	if err := r.Run("127.0.0.1:" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

// seed 首次启动写入演示数据（仅当用户表为空）。
func seed() {
	var n int64
	db.Model(&User{}).Count(&n)
	if n > 0 {
		return
	}
	db.Create(&User{Email: "admin@yunyue.cn", PasswordHash: hashPassword("admin12345"), Role: "admin", Name: "店长"})
	db.Create(&User{Email: "staff@yunyue.cn", PasswordHash: hashPassword("staff12345"), Role: "staff", Name: "小雪"})
	db.Create(&[]ServiceItem{
		{Name: "精剪", DurationMin: 45, PriceCents: 6800, Active: true},
		{Name: "染发", DurationMin: 90, PriceCents: 28800, Active: true},
		{Name: "头皮护理", DurationMin: 60, PriceCents: 16800, Active: true},
	})
	db.Create(&[]Staff{
		{Name: "阿明", Phone: "13700000001", Title: "发型总监", Active: true},
		{Name: "小雪", Phone: "13700000002", Title: "高级发型师", Active: true},
		{Name: "强子", Phone: "13700000003", Title: "护理技师", Active: true},
	})
	db.Create(&Member{Name: "张三", Phone: "13800000001", BalanceCents: 50000, DiscountPerson: 90, CreatedAt: time.Now().Unix()})
	log.Println("seeded demo data")
}

// CORS 开发期前端 vite 直连；生产走 nginx 同源，不受影响。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
