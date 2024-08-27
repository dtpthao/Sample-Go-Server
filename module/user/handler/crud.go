package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sample-go-server/entity"
	"sample-go-server/utils"
)

func (h UserHandler) CreateUser(c *gin.Context) {
	var userReq entity.NewUserRequest
	err := c.ShouldBindJSON(&userReq)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	// todo validate input
	newUser := entity.User{
		Uuid:     utils.NewUuid(),
		Username: userReq.Username,
		Password: userReq.Password, // todo hash password from frontend?
		IsAdmin:  userReq.IsAdmin,
	}

	user, err := h.uc.Create(newUser)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h UserHandler) GetListUsers(c *gin.Context) {
	res, err := h.uc.GetList()
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h UserHandler) UpdateUser(c *gin.Context) {

	uUuid := c.Param("uuid")

	var uReg entity.UpdateUserRequest
	err := c.ShouldBindJSON(&uReg)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err)
		return
	}

	// todo validate input
	user := entity.User{
		Uuid:     uUuid,
		Username: uReg.Username,
		Password: uReg.Password,
		IsAdmin:  uReg.IsAdmin,
	}

	err = h.uc.Update(user)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h UserHandler) GetUserInfo(c *gin.Context) {

	uUuid := c.Param("uuid")
	user, err := h.uc.GetUserByUuid(uUuid)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err) // fixme error can be either BadRequest or Internal
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h UserHandler) DeleteUser(c *gin.Context) {
	uUid := c.Param("uuid")
	err := h.uc.Delete(uUid)
	if err != nil {
		utils.HandleError(c, http.StatusBadRequest, err) // fixme status can be either BadRequest or Internal error
		return
	}

	c.JSON(http.StatusNoContent, entity.DeleteUserResponse{Success: true})
}
