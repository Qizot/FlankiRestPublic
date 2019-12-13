package authorization

import (
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID 		 uint
	Email    string
	Password string
}

func FetchUserID(db *gorm.DB, email, password string) (id uint, err error) {
	user := &User{}
	if err = db.DB().Ping(); err != nil {
		err = ErrDatabaseError
		return
	}
	err = db.Table("accounts").Where("email=?", email).Scan(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = ErrUnauthorizedAccount
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			err = ErrUnauthorizedAccount
		} else {
			err = ErrCrypto
		}
		return
	}
	id = user.ID
	return
}
