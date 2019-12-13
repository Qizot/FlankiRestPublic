package services

import (
	"FlankiRest/errors"
	"FlankiRest/logger"
	"FlankiRest/models"
	"bytes"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"os"
	"time"
)

type ResetModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
}

type PasswordReset struct {
	ResetModel
	AccountID uint
	Code      string
}

type PasswordResetTemplate struct {
	Nickname string
	URL      string
}

type ResetRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

func EmailPasswordResetRequest(db *gorm.DB, email string) error {

	account := &models.Account{}
	err := db.Model(account).Where("email = ?", email).Select("id, nickname, email").First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("Email '"+email+"' has not been found", 400)
		}
		return errors.DatabaseError(err)
	}

	resetCode, err := uuid.NewRandom()
	if err != nil {
		return errors.New("Error while generating uuid: "+err.Error(), 500)
	}

	resetEntry := &PasswordReset{}
	resetEntry.AccountID = account.ID
	resetEntry.Code = resetCode.String()

	err = db.Create(&resetEntry).Error
	if err != nil {
		return errors.DatabaseError(err)
	}
	domain, found := os.LookupEnv("RESET_PASSWORD_DOMAIN")
	if !found {
		return errors.New("There is no front end domain to which redirect the user", 500)
	}
	data := PasswordResetTemplate{account.Nickname, domain + "/" + resetEntry.Code}

	tmpl, err := template.ParseFiles("./templates/passwordReset.txt")
	if err != nil {
		return errors.New("Error while parsing template: "+err.Error(), 500)
	}
	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, data)
	if err != nil {
		return errors.New("Error while filling template: "+err.Error(), 500)
	}

	mailAuth := GetAppMailAuth()
	sender := EmailSender{To: []string{email}, Subject: "Password reset", Body: &buffer}

	go func() {
		err = mailAuth.SendEmail(sender, true)
		if err != nil {
			logger.GetGlobalLogger().WithField("prefix", "[EMAIL SERVICE]").Error(err.Error())
		}
	}()
	return nil
}

func ResetPassword(db *gorm.DB, request ResetRequest) error {
	resetEntry := &PasswordReset{}
	allowed_time := time.Now().Add(time.Minute * time.Duration(-15))
	err := db.Model(resetEntry).Where("code = ? and created_at > ?", request.Code, allowed_time).First(resetEntry).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("code has not been found or is already expired", 400)
		}
		return errors.DatabaseError(err)
	}

	account := &models.Account{}
	err = db.Model(account).Where("id = ?", resetEntry.AccountID).First(account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("account has not been found", 404)
		}
		return errors.DatabaseError(err)
	}

	if size := len(request.NewPassword); size < 6 || size > 32 {
		return errors.New("Invalid password size, range 6 to 32", 400)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	account.Password = string(hashedPassword)
	err = db.Save(account).Error
	if err != nil {
		return errors.DatabaseError(err)
	}
	err = db.Delete(resetEntry).Error
	return err


}

