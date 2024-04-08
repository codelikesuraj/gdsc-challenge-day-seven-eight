package controllers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/codelikesuraj/gdsc-challenge-seven-eight/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (b *UserController) Register(c *gin.Context) {
	var count int64
	var user models.User

	if err := c.ShouldBind(&user); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid input",
				"errors":  models.GetValidationErrs(ve),
			})
			return
		}

		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
			"errors":  nil,
		})
		return
	}

	// check if user exists
	result := b.DB.Model(&models.User{}).Where("username = ?", user.Username).Count(&count)
	switch {
	case result.Error != nil:
		log.Println(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	case count > 0:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "user already exists"})
		return
	}

	// hash password
	if err := user.HashPassword(user.Password); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	// create user
	result = b.DB.Create(&user)
	if result.Error != nil || result.RowsAffected < 1 {
		log.Println(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

func (b *UserController) Login(c *gin.Context) {
	var user models.User

	if err := c.ShouldBind(&user); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"message": "invalid input",
				"errors":  models.GetValidationErrs(ve),
			})
			return
		}

		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
			"errors":  nil,
		})
		return
	}

	username, password := user.Username, user.Password

	// hash password
	if err := user.HashPassword(password); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	// check if user exists
	result := b.DB.Where("username = ?", username).First(&user)
	switch {
	case result.Error == gorm.ErrRecordNotFound:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid credentials"})
		return
	case result.Error != nil:
		log.Println(result.Error)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	// check password
	err := user.CheckPassword(password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "invalid credentials"})
		return
	}

	// generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 10).Unix(),
	})

	// sign and get the complete encoded token as a string using the secret key
	tokenString, err := token.SignedString([]byte("SECRET_KEY"))
	if err != nil {
		log.Println(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600, "", "", false, true)

	c.JSON(http.StatusOK, gin.H{})
}

func (b *UserController) Validate(c *gin.Context) {
	auth, _ := c.Get("auth_id")
	c.JSON(http.StatusOK, gin.H{
		"message": "I am logged in!",
		"data":    auth,
	})
}
