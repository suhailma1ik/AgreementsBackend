package handler

import (
	"fmt"
	"main/config"
	"main/database"
	"main/model"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims := t.Claims.(jwt.MapClaims)
	uid := int(claims["user_id"].(float64))

	return uid == n
}

func validUser(id string, p string) bool {
	db := database.DB
	var user model.User
	db.First(&user, id)
	if user.Username == "" {
		return false
	}
	if !CheckPasswordHash(p, user.Password) {
		return false
	}
	return true
}

// GetUser get a user
func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DB
	var user model.User
	db.Find(&user, id)
	if user.Username == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No user found with ID", "data": nil})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "User found", "data": user})
}



// CreateUser new user
func CreateUser(c *fiber.Ctx) error {
	fmt.Println("create user")
	type NewUser struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
	    CurrentSerialNumber int `json:"current_serial_number"`
		Id int `json:"id"`
	}
	db := database.DB
	user := new(model.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	hash, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't hash password", "data": err})

	}

	user.Password = hash
	if err := db.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "data": err})
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["user_id"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() // Token expiration time

	// Sign the token with a secret key
	t, err := token.SignedString([]byte(config.Config("SECRET")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't generate JWT token", "data": err})
	}

	newUser := NewUser{
		Email:    user.Email,
		Username: user.Username,
		Phone:    user.Phone,
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Created user", "data": newUser, "token": t,"id":user.ID})
}

// // UpdateUser update user
// func UpdateUser(c *fiber.Ctx) error {
// 	type UpdateUserInput struct {
// 		Names string `json:"names"`
// 	}
// 	var uui UpdateUserInput
// 	if err := c.BodyParser(&uui); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
// 	}
// 	id := c.Params("id")
// 	token := c.Locals("user").(*jwt.Token)

// 	if !validToken(token, id) {
// 		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Invalid token id", "data": nil})
// 	}

// 	db := database.DB
// 	var user model.User

// 	db.First(&user, id)
// 	user.Names = uui.Names
// 	db.Save(&user)

// 	return c.JSON(fiber.Map{"status": "success", "message": "User successfully updated", "data": user,"token":token})
// }

// DeleteUser delete user
func DeleteUser(c *fiber.Ctx) error {
	type PasswordInput struct {
		Password string `json:"password"`
	}
	var pi PasswordInput
	if err := c.BodyParser(&pi); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "data": err})
	}
	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)
	if !validToken(token, id) {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Invalid token id", "data": nil})

	}

	if !validUser(id, pi.Password) {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Not valid user", "data": nil})

	}

	db := database.DB
	var user model.User

	db.First(&user, id)

	db.Delete(&user)
	return c.JSON(fiber.Map{"status": "success", "message": "User successfully deleted", "data": nil})
}



func UpdateSerialNumber(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DB
	var user model.User
	db.First(&user, id)
	user.CurrentSerialNumber = user.CurrentSerialNumber + 1
	db.Save(&user)
	return c.JSON(fiber.Map{"status": "success", "message": "User successfully updated", "data": user})
}