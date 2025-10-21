package handler

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"time"

	middleware "chatbot/internal/controller/http/middlerware"
	"chatbot/internal/controller/http/token"
	"chatbot/internal/entity"
	"chatbot/pkg/cache"

	"github.com/gin-gonic/gin"
)

const phoneRegex = `^(\+998)?[0-9]{9}$`

func isValidPhone(phone string) bool {
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

// // Register godoc
// // @Summary Create a new user
// // @Description Create a new user with the provided details
// // @Tags Users
// // @Accept  json
// // @Produce  json
// // @Param user body entity.CreateUser true "User Details"
// // @Success 200 {object} string
// // @Failure 400 {object}  string
// // @Failure 500 {object} string
// // @Security BearerAuth
// // @Router /users/register [post]
// func (h *Handler) Register(c *gin.Context) {
// 	reqBody := entity.CreateUser{}
// 	err := c.BindJSON(&reqBody)
// 	if err != nil {
// 		c.JSON(400, gin.H{"Error binding request body": err.Error()})
// 		slog.Error("Error binding request body: ", "err", err)
// 		return
// 	}

// 	if !isValidPhone(reqBody.PhoneNumber) {
// 		c.JSON(409, gin.H{"message": "Incorrect phone number format"})
// 		slog.Error("Incorrect phone number format")
// 		return
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		c.JSON(409, gin.H{"error": "Server error"})
// 		slog.Error("Error hashing password: ", "err", err)
// 		return
// 	}
// 	reqBody.Password = string(hashedPassword)

// 	_, err = h.UseCase.UserRepo.Create(context.Background(), &reqBody)
// 	if err != nil {
// 		c.JSON(500, gin.H{"Error creating user:": err.Error()})
// 		slog.Error("Error creating user: ", "err", err)
// 		return
// 	}

// 	slog.Info("New user created successfully")
// 	c.JSON(200, gin.H{"Massage": "User registered successfully"})
// }

// Login godoc
// @Summary User login
// @Description User login with phone number and password
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body entity.LoginReq true "User Login Details"
// @Success 200 {object} entity.LoginRes
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /users/login [post]
func (h *Handler) Login(c *gin.Context) {
	reqBody := entity.LoginReq{}
	err := c.BindJSON(&reqBody)
	if err != nil {
		c.JSON(400, gin.H{"Error binding request body": err.Error()})
		slog.Error("Error binding request body: ", "err", err)
		return
	}

	if !isValidPhone(reqBody.PhoneNumber) {
		c.JSON(409, gin.H{"message": "Incorrect phone number format"})
		slog.Error("Incorrect phone number format")
		return
	}

	code, err := generateVerificationCode()
	if err != nil {
		c.JSON(500, gin.H{"Error generating verification code:": err.Error()})
		slog.Error("Error generating verification code: ", "err", err)
		return
	}

	// go helper.SendSms(*h.Config, reqBody.PhoneNumber, code)

	go cache.SaveVerificationCode(h.Redis, context.Background(), reqBody.PhoneNumber, code, 3*time.Minute)

	exist, err := h.UseCase.UserRepo.CheckExist(context.Background(), reqBody.PhoneNumber)
	if err != nil {
		c.JSON(500, gin.H{"Error checking user existence:": err.Error()})
		slog.Error("Error checking user existence: ", "err", err)
		return
	}

	if !exist {
		_, err := h.UseCase.UserRepo.Create(context.Background(), &entity.CreateUser{
			PhoneNumber: reqBody.PhoneNumber,
		})
		if err != nil {
			c.JSON(500, gin.H{"Error creating user:": err.Error()})
			slog.Error("Error creating user: ", "err", err)
			return
		}
		slog.Info("New user created successfully")
		c.JSON(200, gin.H{"code": code})
		return
	} else {
		c.JSON(200, gin.H{"code": code})
		return
	}
}

// Verify godoc
// @Summary Verify user login
// @Description Verify user by SMS code
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body entity.VerifyReq true "User Verify Details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/verify [post]
func (h *Handler) Verify(c *gin.Context) {
	var req entity.VerifyReq

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	fmt.Println(req)
	storedCode, err := cache.GetVerificationCode(h.Redis, context.Background(), req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Verification code expired or not found"})
		slog.Error("Error getting verification code", "err", err)
		return
	}

	if storedCode != req.Code {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect verification code"})
		return
	}

	user, err := h.UseCase.UserRepo.GetByPhone(context.Background(), req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
		slog.Error("Error retrieving user by phone: ", "err", err)
		return
	}

	tokenStr := token.GenerateJWTToken(user.Id, user.Role)

	c.SetCookie(
		"access_token",
		tokenStr.AccessToken,
		3600,
		"",
		"",
		false,
		true,
	)

	fmt.Println("Generated Access Token:", tokenStr.AccessToken)

	go cache.DeleteVerificationCode(h.Redis, context.Background(), req.PhoneNumber)

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification successful",
	})
}

// GetByIdUser godoc
// @Summary Get User by ID
// @Description Get a User by their ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Success 200 {object} entity.UserInfo
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /users/profile [get]
func (h *Handler) GetByIdUser(c *gin.Context) {

	claims, err := middleware.ExtractToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Access token missing or invalid"})
		return
	}

	userID, ok := claims["id"].(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	res, err := h.UseCase.UserRepo.GetById(context.Background(), &entity.ById{Id: userID})
	if err != nil {
		c.JSON(500, gin.H{"Error getting User by ID: ": err.Error()})
		slog.Error("Error getting User by ID: ", "err", err)
		return
	}

	slog.Info("User retrieved successfully")
	c.JSON(200, res)
}

// UpdateUser godoc
// @Summary Update a User
// @Description Update a User's details
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id query string true "User ID"
// @Param User body entity.UpdateUserBody true "User Update Details"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /users/update [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	reqBody := entity.UpdateUserBody{}

	err := c.BindJSON(&reqBody)
	if err != nil {
		c.JSON(400, gin.H{"Error binding request body:": err.Error()})
		slog.Error("Error binding request body: ", "err", err)
		return
	}

	err = h.UseCase.UserRepo.Update(context.Background(), &entity.UpdateUser{
		Id:          c.Query("id"),
		FullName:    reqBody.FullName,
		PhoneNumber: reqBody.PhoneNumber,
	})
	if err != nil {
		c.JSON(500, gin.H{"Error updating User:": err.Error()})
		slog.Error("Error updating User: ", "err", err)
		return
	}

	slog.Info("User updated successfully")
	c.JSON(200, "User updated successfully")
}

// GetAllUsers godoc
// @Summary Get all Users
// @Description Get all Users with optional filtering
// @Tags Users
// @Accept  json
// @Produce  json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param status query string false "User Status"
// @Success 200 {object} entity.UserList
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /users/list [get]
func (h *Handler) GetAllUsers(c *gin.Context) {
	limit := c.Query("limit")
	offset := c.Query("offset")
	status := c.Query("status")

	limitValue, offsetValue, err := parsePaginationParams(c, limit, offset)
	if err != nil {
		c.JSON(400, gin.H{"Error parsing pagination parameters:": err.Error()})
		slog.Error("Error parsing pagination parameters: ", "err", err)
		return
	}

	req := &entity.Filter{
		Limit:  limitValue,
		Offset: offsetValue,
	}

	res, err := h.UseCase.UserRepo.GetAll(context.Background(), req, status)
	if err != nil {
		c.JSON(500, gin.H{"Error getting Users:": err.Error()})
		slog.Error("Error getting Users: ", "err", err)
		return
	}

	slog.Info("Users retrieved successfully")
	c.JSON(200, res)
}

// DeleteUser godoc
// @Summary Delete a User
// @Description Delete a User by ID
// @Tags Users
// @Accept  json
// @Produce  json
// @Param id query string true "User ID"
// @Success 200 {string} string "User deleted successfully"
// @Failure 400 {object} string
// @Failure 500 {object} string
// @Security BearerAuth
// @Router /users/delete [delete]
func (h *Handler) DeleteUser(c *gin.Context) {

	err := h.UseCase.UserRepo.Delete(context.Background(), &entity.ById{Id: c.Query("id")})
	if err != nil {
		c.JSON(500, gin.H{"Error deleting User by ID:": err.Error()})
		slog.Error("Error deleting User by ID: ", "err", err)
		return
	}

	slog.Info("User deleted successfully")
	c.JSON(200, "User deleted successfully")
}

// Logout godoc
// @Summary User logout
// @Description Clear user cookie
// @Tags Users
// @Success 200 {object} string
// @Router /users/logout [post]
func (h *Handler) Logout(c *gin.Context) {

	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}

func parsePaginationParams(c *gin.Context, limit, offset string) (int, int, error) {
	limitValue := 10
	offsetValue := 0

	if limit != "" {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil {
			slog.Error("Invalid limit value", "err", err.Error())
			c.JSON(400, gin.H{"error": "Invalid limit value"})
			return 0, 0, err
		}
		limitValue = parsedLimit
	} else {
		limitValue = 0
	}

	if offset != "" {
		parsedOffset, err := strconv.Atoi(offset)
		if err != nil {
			slog.Error("Invalid offset value", "err", err.Error())
			c.JSON(400, gin.H{"error": "Invalid offset value"})
			return 0, 0, err
		}
		offsetValue = parsedOffset
	}

	return limitValue, offsetValue, nil
}

func generateVerificationCode() (string, error) {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(899999) + 100000
	return fmt.Sprintf("%06d", code), nil
}
