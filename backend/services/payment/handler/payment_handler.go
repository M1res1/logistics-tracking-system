package handler

import (
    "net/http"
    "strconv"

    "logistics-tracking-system/pkg/response"
    "logistics-tracking-system/services/payment/service"

    "github.com/gin-gonic/gin"
)

type PaymentHandler struct {
    svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
    return &PaymentHandler{svc: svc}
}

type processPaymentReq struct {
    OrderID        uint    `json:"order_id"`
    UserID         uint    `json:"user_id"`
    Amount         float64 `json:"amount"`
    Method         string  `json:"method"`
    IdempotencyKey string  `json:"idempotency_key"`
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
    var req processPaymentReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request body")
        return
    }

    p, err := h.svc.ProcessPayment(c.Request.Context(), &service.ProcessPaymentRequest{
        OrderID:        req.OrderID,
        UserID:         req.UserID,
        Amount:         req.Amount,
        Method:         req.Method,
        IdempotencyKey: req.IdempotencyKey,
    })
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    response.Success(c, p)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid id")
        return
    }

    p, err := h.svc.GetPayment(uint(id))
    if err != nil {
        response.Error(c, http.StatusNotFound, "payment not found")
        return
    }

    response.Success(c, p)
}

type refundReq struct {
    Amount float64 `json:"amount"`
    Reason string  `json:"reason"`
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid id")
        return
    }

    var req refundReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request body")
        return
    }

    ref, err := h.svc.RefundPayment(c.Request.Context(), uint(id), req.Amount, req.Reason)
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    response.Success(c, ref)
}

func (h *PaymentHandler) GetWallet(c *gin.Context) {
    userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid user id")
        return
    }

    w, err := h.svc.GetWallet(uint(userID))
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    response.Success(c, w)
}

type topupReq struct {
    Amount float64 `json:"amount"`
}

func (h *PaymentHandler) TopupWallet(c *gin.Context) {
    userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "invalid user id")
        return
    }

    var req topupReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "invalid request body")
        return
    }

    w, err := h.svc.TopupWallet(uint(userID), req.Amount)
    if err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    response.Success(c, w)
}