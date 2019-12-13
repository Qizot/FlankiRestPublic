package models

import (
	"FlankiRest/errors"
	"encoding/json"
	"fmt"
	"github.com/badoux/checkmail"
	"github.com/evanphx/json-patch"
	"reflect"
	"time"

	//"strings"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

type AccountModel struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}


type Account struct {
	AccountModel
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"`
	Sex         string `json:"sex"`
	Description string `json:"description"`
	Playing     bool   `json:"playing"`
}

type UpdateAccount struct {
	ID          uint
	Nickname    string          `json:"nickname,omitempty"`
	Email       string          `json:"email,omitempty"`
	Password    string          `json:"password,omitempty"`
	Sex         string          `json:"sex,omitempty"`
	Description json.RawMessage `json:"description,omitempty"` // to enable updating description to be empty again json has to distinguish between empty string vs non existent, here empty string is of length 2 and non-set is 0
}

// To enable comparing Description json.RawMessage reflect must be used because slices (here RawMessage) are not comparable in golang
func (m UpdateAccount ) IsEmpty() bool {
	return reflect.DeepEqual(UpdateAccount{}, m)
}

func (account *Account) ValidateFieldsRequirements() error {

	if size := len(account.Nickname); size < 4 || size > 20 {
		return errors.New("Wrong nickname size, should be from 4 to 20 letters", 400)
	}

	if account.Sex != "male" && account.Sex != "female" {
		return errors.New("Unknown sex, should be either male or female", 400)
	}

	if err := checkmail.ValidateFormat(account.Email); err != nil {
		return errors.New("Incorrect email format", 400)
	}

	if size := len(account.Password); size < 6 || size > 32 {
		if size < 6 {
			return errors.New("Password is too short(6 to 32 signs)", 400)
		} else {
			return errors.New("Password is too long(6 to 32 signs)", 400)
		}
	}

	if len(account.Description) > 200 {
		return errors.New("Description is too long, up to signs", 400)
	}
	return nil
}

func IsAccountFieldUnique(db *gorm.DB, fieldName string, fieldValue string) error {
	var count int
	account := &Account{}
	err := db.Model(&account).Where(fmt.Sprintf("%s = ?", fieldName), fieldValue).Count(&count).Error
	if err != nil {
		return errors.New(fmt.Sprintf("database connection error: %s", err.Error()), 500)
	}
	if count != 0 {
		return errors.New(fmt.Sprintf("Field '%s' is not unique", fieldName), 400)
	}
	return nil
}

func (account *Account) Validate(db *gorm.DB) error {

	err := account.ValidateFieldsRequirements()
	if err != nil {
		return err
	}

	err = IsAccountFieldUnique(db,"nickname", account.Nickname)
	if err != nil {
		return err
	}

	err= IsAccountFieldUnique(db,"email", account.Email)
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) Create(db *gorm.DB) error {

	err := account.Validate(db)
	if err != nil {
		return err
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)

	db.Create(&account)
	if account.ID <= 0 {
		return errors.New("Failed to create account, connection error.", 500)
	}
	return nil
}

func (account *Account) Delete(db *gorm.DB) error {
	err := db.Delete(account).Error
	if err != nil {
		return errors.New("Encountered error in database while trying to delete account", 500)
	}
	return nil
}

/*
	To prevent request from changing certain fields request is unmarshaled to temporary UpdateAccount struct
	containing only allowed fields to be changed.
	We then marshal actual user's account to json and request's data to another json.
	JSON patch ( RFC6902 JSON patches ) is performed to update account's
	fields with the ones included in update request.
	Finally, patched json is marshalled back to real account struct, validated and eventually
	updated to the database.

 */

func (toUpdate *UpdateAccount) Update(db *gorm.DB) error {

	account := &Account{}
	err := db.First(&account, toUpdate.ID).Error
	if err == gorm.ErrRecordNotFound {
		return errors.New("Account has not been found", 400)
	}
	if toUpdate.Nickname != "" && toUpdate.Nickname != account.Nickname {
		err = IsAccountFieldUnique(db,"nickname", toUpdate.Nickname)
		if err != nil {
			return err
		}
	}
	if toUpdate.Email != "" && toUpdate.Email != account.Email {
		err = IsAccountFieldUnique(db,"email", toUpdate.Email)
		if err != nil {
			return err
		}
	}
	if toUpdate.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(toUpdate.Password), bcrypt.DefaultCost)
		toUpdate.Password = string(hashedPassword)
	}

	updateJson, _ := json.Marshal(toUpdate)
	accountJson, _ := json.Marshal(account)
	newJson, _ := jsonpatch.MergePatch(accountJson, updateJson)
	err = json.Unmarshal(newJson, account)

	if err != nil {
		return errors.New("Failed to update account information", 500)
	}

	hash := account.Password
	account.Password = "XXXXXX" // nasty trick to pass password field validation
	err = account.ValidateFieldsRequirements()
	if err != nil {
		return err
	}
	account.Password = hash
	db.Save(account)

	return nil
}

func GetAccountById(db *gorm.DB, id uint) (*Account, error) {
	account := &Account{}
	err := db.Where("id = ?", id).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New("Account has not been found", 400)
		} else {
			err = errors.New(fmt.Sprintf("Database error: %s", err.Error()), 500)
		}
		return nil, err
	}
	return account, nil
}

