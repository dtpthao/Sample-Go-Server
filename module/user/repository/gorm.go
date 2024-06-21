package repository

import (
	"errors"
	"glintecoTask/entity"
	"gorm.io/gorm"
	"time"
)

const TableUsers = "users"

type UserRepository struct {
	Uuid      string    `gorm:"primaryKey;column:uuid"`
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	IsAdmin   bool      `gorm:"column:is_admin"`
	Active    bool      `gorm:"column:active"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoCreateTime;column:updated_at"`
	db        *gorm.DB
}

func (r UserRepository) FromEntity(u entity.User) UserRepository {
	return UserRepository{
		Uuid:     u.Uuid,
		Username: u.Username,
		Password: u.Password,
		IsAdmin:  u.IsAdmin,
		Active:   u.Active,
	}
}

func NewUserRepository(db *gorm.DB) entity.IUserRepo {
	return &UserRepository{db: db}
}

func (r UserRepository) ToEntity() *entity.User {
	return &entity.User{
		Uuid:     r.Uuid,
		Username: r.Username,
		Password: r.Password,
		IsAdmin:  r.IsAdmin,
		Active:   r.Active,
	}
}

func (r UserRepository) Create(u entity.User) (*entity.User, error) {
	ur := r.FromEntity(u)
	err := r.db.Table(TableUsers).Create(&ur).Error
	return ur.ToEntity(), err
}

func (r UserRepository) GetUserByUsername(username string) (*entity.User, error) {
	var res UserRepository
	err := r.db.Table(TableUsers).Where("username = ?", username).Take(&res).Error
	if err != nil {
		return nil, err
	}
	return res.ToEntity(), nil
}

func (r UserRepository) GetUserByUuid(uuid string) (*entity.User, error) {
	var res UserRepository
	err := r.db.Table(TableUsers).Where("uuid = ?", uuid).Take(&res).Error
	if err != nil {
		return nil, err
	}
	return res.ToEntity(), nil
}

func (r UserRepository) Delete(username string) error {
	res := r.db.Model(&UserRepository{}).Where("username = ?", username).Set("active", false)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return errors.New("failed to delete user")
	}

	return nil
}

func (r UserRepository) GetList() ([]entity.User, error) {
	var users []entity.User
	err := r.db.Table(TableUsers).Find(&users).Error
	return users, err
}

func (r UserRepository) Update(uuid string, data map[string]any) error {

	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	res := tx.Table(TableUsers).Where("uuid = ?", uuid).Updates(data)
	if res.Error != nil {
		tx.Rollback()
		return res.Error
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return errors.New("failed to update user")
	}

	if tx.Commit().Error != nil {
		tx.Rollback()
		return errors.New("cannot commit transaction")
	}

	return nil
}
