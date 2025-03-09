package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Loan Model
type Loan struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	BorrowerID      string    `json:"borrowerId"`
	Principal       float64   `json:"principalAmount"`
	Rate            float64   `json:"rate"`
	State           string    `json:"state" gorm:"default:proposed"`
	ApprovalDate    time.Time `json:"approvalDate,omitempty"`
	ProofImageURL   string    `json:"proofImageUrl,omitempty"`
	FieldValidator  string    `json:"fieldValidatorId,omitempty"`
	AgreementURL    string    `json:"agreementLetterUrl,omitempty"`
	DisbursementDate time.Time `json:"disbursementDate,omitempty"`
	FieldOfficerID  string    `json:"fieldOfficerId,omitempty"`
}

// LoanService struct
type LoanService struct {
	DB *gorm.DB
}

// NewLoanService initializes the service
func NewLoanService(db *gorm.DB) *LoanService {
	return &LoanService{DB: db}
}

// Create Loan Handler
func (s *LoanService) CreateLoanHandler(c *gin.Context) {
	var loan Loan
	if err := c.ShouldBindJSON(&loan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.DB.Create(&loan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create loan"})
		return
	}
	c.JSON(http.StatusCreated, loan)
}

// Approve Loan Handler
func (s *LoanService) ApproveLoanHandler(c *gin.Context) {
	id := c.Param("id")
	var loan Loan
	if err := s.DB.First(&loan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}
	if loan.State != "proposed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in proposed state"})
		return
	}

	var request struct {
		ProofImageURL  string    `json:"proofImageUrl"`
		FieldValidator string    `json:"fieldValidatorId"`
		ApprovalDate   time.Time `json:"approvalDate"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan.State = "approved"
	loan.ProofImageURL = request.ProofImageURL
	loan.FieldValidator = request.FieldValidator
	loan.ApprovalDate = request.ApprovalDate

	s.DB.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

// Invest Loan Handler
func (s *LoanService) InvestLoanHandler(c *gin.Context) {
	id := c.Param("id")
	var loan Loan
	if err := s.DB.First(&loan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}
	if loan.State != "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in approved state"})
		return
	}

	var request struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if loan.Principal < request.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Investment exceeds principal"})
		return
	}

	loan.State = "invested"
	s.DB.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

// Disburse Loan Handler
func (s *LoanService) DisburseLoanHandler(c *gin.Context) {
	id := c.Param("id")
	var loan Loan
	if err := s.DB.First(&loan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}
	if loan.State != "invested" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Loan must be in invested state"})
		return
	}

	var request struct {
		AgreementURL    string    `json:"agreementLetterUrl"`
		FieldOfficerID  string    `json:"fieldOfficerId"`
		DisbursementDate time.Time `json:"disbursementDate"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan.State = "disbursed"
	loan.AgreementURL = request.AgreementURL
	loan.FieldOfficerID = request.FieldOfficerID
	loan.DisbursementDate = request.DisbursementDate

	s.DB.Save(&loan)
	c.JSON(http.StatusOK, loan)
}

// Get Loan ROI Handler
func (s *LoanService) GetLoanROIHandler(c *gin.Context) {
	id := c.Param("id")
	var loan Loan
	if err := s.DB.First(&loan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	// Calculate ROI (assuming simple interest)
	timeDuration := time.Since(loan.ApprovalDate).Hours() / (24 * 365) // Convert to years
	totalInterest := loan.Principal * loan.Rate * timeDuration
	roi := (totalInterest / loan.Principal) * 100

	c.JSON(http.StatusOK, gin.H{"loanId": loan.ID, "ROI": roi})
}
