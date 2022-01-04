package core

import (
	"fmt"

	"github.com/delta/orientation-backend/models"
	"github.com/labstack/echo/v4"
)

type userUpdateRequest struct {
	User struct {
		Email        string        `json:"email"`
		Name         string        `json:"name"`
		RefreshToken string        `json:"refreshToken"`
		Gender       models.Gender `json:"gender"`
		Department   string        `json:"department" validate:"required"`
		Username     string        `json:"username" validate:"required"`
		Description  string        `json:"description"`
		SpriteType   string        `json:"spriteType"`
	} `json:"user"`
}

func newUserUpdateRequest() *userUpdateRequest {
	return new(userUpdateRequest)
}

// Adds user data to userUpdateRequest
func (r *userUpdateRequest) populate(u *models.User) {

	r.User.Email = u.Email
	r.User.Name = u.Name
	r.User.RefreshToken = u.RefreshToken
	r.User.Gender = u.Gender
	r.User.Department = u.Department

	if u.Username != "" {
		r.User.Username = u.Username
	}
	if u.Description != "" {
		r.User.Description = u.Description
	}
}

// Adds request data to user model and validates it
func (r *userUpdateRequest) bind(c echo.Context, u *models.User) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	fmt.Printf("%+v", r)
	if err := c.Validate(r); err != nil {
		return err
	}
	u.Username = r.User.Username
	u.Description = r.User.Description
	u.Gender = r.User.Gender
	u.Department = r.User.Department
	u.SpriteType = r.User.SpriteType

	return nil
}

// func (r *userUpdateRequest) bind(c echo.Context, u *model.User) error {
// 	if err := c.Bind(r); err != nil {
// 		return err
// 	}
// 	if err := c.Validate(r); err != nil {
// 		return err
// 	}
// 	u.Username = r.User.Username
// 	u.Email = r.User.Email
// 	if r.User.Password != u.Password {
// 		h, err := u.HashPassword(r.User.Password)
// 		if err != nil {
// 			return err
// 		}
// 		u.Password = h
// 	}
// 	u.Bio = &r.User.Bio
// 	u.Image = &r.User.Image
// 	return nil
// }
