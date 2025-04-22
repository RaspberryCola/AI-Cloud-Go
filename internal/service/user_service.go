package service

import (
	"ai-cloud/config"
	"ai-cloud/internal/dao"
	"ai-cloud/internal/middleware"
	"ai-cloud/internal/model"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(user *model.User) error
	Login(req *model.UserNameLoginReq) (*model.LoginResponse, error)
}

type userService struct {
	userDao dao.UserDao
}

func NewUserService(userDao dao.UserDao) UserService {
	return &userService{userDao: userDao}
}

func (s *userService) Register(user *model.User) error {
	// 检查
	usernameExists, err := s.userDao.CheckFieldExists("username", user.Username)
	if err != nil {
		return err
	}
	if usernameExists {
		return errors.New("用户名已注册")
	}

	phoneExists, err := s.userDao.CheckFieldExists("phone", user.Phone)
	if err != nil {
		return err
	}
	if phoneExists {
		return errors.New("手机号已注册")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	newUser := &model.User{
		Username: user.Username,
		Phone:    user.Phone,
		Password: string(hashedPassword),
		Email:    user.Email,
	}
	err = s.userDao.CreateUser(newUser)
	if err != nil {
		return err
	}
	return nil
}

func (s *userService) Login(req *model.UserNameLoginReq) (*model.LoginResponse, error) {
	user, err := s.userDao.GetUserByName(req.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	accessToken, err := middleware.GenerateToken(user.ID)
	if err != nil {
		return nil, errors.New("系统错误")
	}

	return &model.LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   config.AppConfigInstance.JWT.ExpirationHours * 3600,
		TokenType:   "Bearer",
	}, nil
}
