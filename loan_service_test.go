package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB initializes an in-memory database for testing
func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}
	db.AutoMigrate(&Loan{})
	return db
}

// setupService initializes LoanService with test DB
func setupService() *LoanService {
	db := setupTestDB()
	return NewLoanService(db)
}

// helper function to perform HTTP requests
func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var jsonBody []byte
	if body != nil {
		jsonBody, _ = json.Marshal(body)
	}
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCreateLoan(t *testing.T) {
	svc := setupService()
	router := gin.Default()
	router.POST("/loans", svc.CreateLoanHandler)

	loan := Loan{
		BorrowerID: "12345",
		Principal:  50000,
		Rate:       5.5,
	}

	w := performRequest(router, "POST", "/loans", loan)
	assert.Equal(t, http.StatusCreated, w.Code)

	var response Loan
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.NotZero(t, response.ID)
	assert.Equal(t, "proposed", response.State)
}

func TestApproveLoan(t *testing.T) {
	svc := setupService()
	router := gin.Default()
	router.POST("/loans", svc.CreateLoanHandler)
	router.POST("/loans/:id/approve", svc.ApproveLoanHandler)

	// Create a loan first
	loan := Loan{
		BorrowerID: "12345",
		Principal:  50000,
		Rate:       5.5,
	}
	createRes := performRequest(router, "POST", "/loans", loan)
	assert.Equal(t, http.StatusCreated, createRes.Code)

	var createdLoan Loan
	json.Unmarshal(createRes.Body.Bytes(), &createdLoan)

	// Approve the loan
	approvalPayload := gin.H{
		"proofImageUrl":  "proof.jpg",
		"fieldValidatorId": "emp123",
		"approvalDate":    time.Now(),
	}
	approveRes := performRequest(router, "POST", "/loans/"+string(rune(createdLoan.ID))+"/approve", approvalPayload)
	assert.Equal(t, http.StatusOK, approveRes.Code)

	// Validate that state is updated
	var approvedLoan Loan
	svc.DB.First(&approvedLoan, createdLoan.ID)
	assert.Equal(t, "approved", approvedLoan.State)
}

func TestInvestLoan(t *testing.T) {
	svc := setupService()
	router := gin.Default()
	router.POST("/loans", svc.CreateLoanHandler)
	router.POST("/loans/:id/approve", svc.ApproveLoanHandler)
	router.POST("/loans/:id/invest", svc.InvestLoanHandler)

	// Create and approve a loan first
	loan := Loan{
		BorrowerID: "12345",
		Principal:  50000,
		Rate:       5.5,
		State:      "approved",
	}
	svc.DB.Create(&loan)

	// Invest in loan
	investmentPayload := gin.H{"amount": 30000}
	investRes := performRequest(router, "POST", "/loans/"+string(rune(loan.ID))+"/invest", investmentPayload)
	assert.Equal(t, http.StatusOK, investRes.Code)

	// Validate that state is updated
	var updatedLoan Loan
	svc.DB.First(&updatedLoan, loan.ID)
	assert.Equal(t, "invested", updatedLoan.State)
}

func TestDisburseLoan(t *testing.T) {
	svc := setupService()
	router := gin.Default()
	router.POST("/loans", svc.CreateLoanHandler)
	router.POST("/loans/:id/disburse", svc.DisburseLoanHandler)

	// Create and approve a loan first
	loan := Loan{
		BorrowerID: "12345",
		Principal:  50000,
		Rate:       5.5,
		State:      "invested",
	}
	svc.DB.Create(&loan)

	// Disburse loan
	disbursePayload := gin.H{
		"agreementLetterUrl": "agreement.pdf",
		"fieldOfficerId":     "emp456",
		"disbursementDate":   time.Now(),
	}
	disburseRes := performRequest(router, "POST", "/loans/"+string(rune(loan.ID))+"/disburse", disbursePayload)
	assert.Equal(t, http.StatusOK, disburseRes.Code)

	// Validate that state is updated
	var updatedLoan Loan
	svc.DB.First(&updatedLoan, loan.ID)
	assert.Equal(t, "disbursed", updatedLoan.State)
}
