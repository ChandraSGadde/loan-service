package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Loan struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	BorrowerID     string    `json:"borrowerId"`
	Principal      float64   `json:"principalAmount"`
	Rate           float64   `json:"rate"`
	ROI            float64   `json:"ROI"`
	State          string    `json:"state" gorm:"default:proposed"`
	ApprovalDate   time.Time `json:"approvalDate,omitempty"`
	ProofImageURL  string    `json:"proofImageUrl,omitempty"`
	FieldValidator string    `json:"fieldValidatorId,omitempty"`
	AgreementURL   string    `json:"agreementLetterUrl,omitempty"`
	DisbursementDate time.Time `json:"disbursementDate,omitempty"`
	FieldOfficerID string    `json:"fieldOfficerId,omitempty"`
}

var db *gorm.DB

func main() {
	r := gin.Default()

	var err error
	db, err = gorm.Open(sqlite.Open("loans.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Loan{})

	r.POST("/loans", createLoan)
	r.POST("/loans/:id/approve", approveLoan)
	r.POST("/loans/:id/invest", investLoan)
	r.POST("/loans/:id/disburse", disburseLoan)

	r.Run(":8080")
}

func createLoan(c *gin.Context) {
	var loan Loan
	if err := c.ShouldBindJSON(&loan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&loan)
	c.JSON(http.StatusCreated, loan)
}

func approveLoan(c *gin.Context) {
	var loan Loan
	id := c.Param("id")
	db.First(&loan, id)
	if loan.State != "proposed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in proposed state"})
		return
	}

	var input struct {
		ProofImageURL  string    `json:"proofImageUrl"`
		FieldValidator string    `json:"fieldValidatorId"`
		ApprovalDate   time.Time `json:"approvalDate"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan.State = "approved"
	loan.ProofImageURL = input.ProofImageURL
	loan.FieldValidator = input.FieldValidator
	loan.ApprovalDate = input.ApprovalDate
	db.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

func investLoan(c *gin.Context) {
	var loan Loan
	id := c.Param("id")
	db.First(&loan, id)
	if loan.State != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in approved state"})
		return
	}

	var input struct {
		InvestorID string  `json:"investorId"`
		Amount     float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if loan.Principal < input.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Investment exceeds principal"})
		return
	}

	loan.State = "invested"
	db.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

func disburseLoan(c *gin.Context) {
	var loan Loan
	id := c.Param("id")
	db.First(&loan, id)
	if loan.State != "invested" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in invested state"})
		return
	}

	var input struct {
		AgreementURL    string    `json:"agreementLetterUrl"`
		FieldOfficerID  string    `json:"fieldOfficerId"`
		DisbursementDate time.Time `json:"disbursementDate"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan.State = "disbursed"
	loan.AgreementURL = input.AgreementURL
	loan.FieldOfficerID = input.FieldOfficerID
	loan.DisbursementDate = input.DisbursementDate
	db.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

