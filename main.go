package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()

	// Initialize Database
	db, err := gorm.Open(sqlite.Open("loans.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&Loan{})

	// Initialize Loan Service
	loanService := NewLoanService(db)

	// Routes
	r.POST("/loans", loanService.CreateLoanHandler)
	r.POST("/loans/:id/approve", loanService.ApproveLoanHandler)
	r.POST("/loans/:id/invest", loanService.InvestLoanHandler)
	r.POST("/loans/:id/disburse", loanService.DisburseLoanHandler)
	r.GET("/loans/:id/roi", loanService.GetLoanROIHandler)

	// Start Server
	r.Run(":8080")
}
