package service

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/savioruz/smrv2-api/internal/dao/entity"
	"github.com/savioruz/smrv2-api/internal/dao/model"
	"github.com/savioruz/smrv2-api/internal/gateway/messaging"
	"github.com/savioruz/smrv2-api/internal/repository"
	"github.com/savioruz/smrv2-api/pkg/helper"
	"github.com/savioruz/smrv2-api/pkg/jwt"
	"github.com/savioruz/smrv2-api/pkg/mail"
	"github.com/savioruz/smrv2-api/pkg/scrape"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

//go:embed template/*.html
var templateFS embed.FS

type UserService interface {
	Register(ctx context.Context, request *model.UsersRegisterRequest) (*model.UsersRegisterResponse, error)
	Login(ctx context.Context, request *model.UsersLoginRequest) (*model.UsersLoginResponse, error)
	RefreshToken(ctx context.Context, request *model.UsersRefreshTokenRequest) (*model.UserRefreshTokenResponse, error)
	VerifyEmail(ctx context.Context, request *model.UsersVerifyEmailRequest) (*model.Response[string], error)
	RequestResetPassword(ctx context.Context, request *model.UserResetPasswordRequest) (*model.Response[string], error)
	ResetPassword(ctx context.Context, request *model.UserResetPassword) (*model.Response[string], error)
}

type UserServiceImpl struct {
	DB                     *gorm.DB
	Log                    *logrus.Logger
	Validator              *validator.Validate
	Viper                  *viper.Viper
	UserRepository         *repository.UserRepositoryImpl
	SubscriptionRepository *repository.SubscriptionRepositoryImpl
	MailProducer           *messaging.MailProducer
	UserProducer           *messaging.UserProducer
	Mail                   *mail.ImplGomail
	Jwt                    *jwt.JWTServiceImpl
}

func NewUserService(
	db *gorm.DB,
	log *logrus.Logger,
	validator *validator.Validate,
	viper *viper.Viper,
	userRepository *repository.UserRepositoryImpl,
	subscriptionRepository *repository.SubscriptionRepositoryImpl,
	mailProducer *messaging.MailProducer,
	userProducer *messaging.UserProducer,
	mail *mail.ImplGomail,
	jwt *jwt.JWTServiceImpl,
) *UserServiceImpl {
	return &UserServiceImpl{
		DB:                     db,
		Log:                    log,
		Validator:              validator,
		Viper:                  viper,
		UserRepository:         userRepository,
		SubscriptionRepository: subscriptionRepository,
		MailProducer:           mailProducer,
		UserProducer:           userProducer,
		Mail:                   mail,
		Jwt:                    jwt,
	}
}

func (s *UserServiceImpl) Register(ctx context.Context, request *model.UsersRegisterRequest) (*model.UsersRegisterResponse, error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	if !strings.Contains(request.Email, "@webmail.uad.ac.id") {
		return nil, helper.SingleError("email", "INVALID_DOMAIN")
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Initialize scraperNewScrape(30)
	scraper := scrape.NewScrape(30)
	if err := scraper.Initialize(); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to initialize scraper")
	}
	defer scraper.Cleanup()

	// Check if user exists
	user, err := s.UserRepository.GetByEmail(tx, request.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, helper.ServerError(s.Log, "Failed to check user existence")
	}
	if user != nil {
		return nil, helper.SingleError("email", "ALREADY_EXISTS")
	}

	// Using encryption because we need to reuse it in the future
	password, err := helper.EncryptPassword(s.Viper, request.Password)
	if err != nil {
		return nil, helper.ServerError(s.Log, err.Error())
	}

	verifyToken := uuid.NewString()

	// Email template preparation
	var replaceEmail = struct {
		Link string
	}{
		Link: fmt.Sprintf("%s/api/v1/auth/verify/%s", s.Viper.GetString("APP_DOMAIN"), verifyToken),
	}

	tmpl, err := template.ParseFS(templateFS, "template/verify_email.html")
	if err != nil {
		s.Log.Errorf("failed to parse template: %v", err)
		return nil, helper.ServerError(s.Log, "Failed to parse template")
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, &replaceEmail); err != nil {
		s.Log.Errorf("failed to execute template: %v", err)
		return nil, helper.ServerError(s.Log, "Failed to execute template")
	}

	// Send email through message queue
	emailMessage := &messaging.EmailMessage{
		To:      request.Email,
		From:    s.Mail.GetFromEmail(),
		Subject: "[Smrv2] Verify your email",
		Body:    body.String(),
	}

	// Create user
	id := uuid.NewString()
	nim := strings.Split(request.Email, "@")[0]
	if err := s.UserRepository.Create(tx, &entity.User{
		ID:                id,
		Email:             request.Email,
		Password:          password,
		Nim:               nim,
		VerificationToken: verifyToken,
		IsVerified:        false,
	}); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to create user")
	}

	if err := s.MailProducer.PublishEmailSending(ctx, emailMessage); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to queue email sending")
	}

	if err := s.UserProducer.PublishStudyDataRequest(ctx, &messaging.StudyDataMessage{
		NIM:       nim,
		Password:  request.Password,
		SendEmail: false,
		UserEmail: "",
		FromEmail: "",
		EmailBody: "",
	}); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to publish study data request")
	}

	if err := s.SubscriptionRepository.Create(tx, &entity.Subscription{
		UserID:    id,
		PlanType:  string(model.SubscriptionPlanFree),
		Status:    string(model.SubscriptionStatusActive),
		StartDate: time.Now().UTC().Add(7 * time.Hour),
		EndDate:   time.Now().UTC().Add(7*time.Hour).AddDate(100, 0, 0),
	}); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to create subscription")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, helper.ServerError(s.Log, "Failed to commit transaction")
	}

	return &model.UsersRegisterResponse{
		Email: request.Email,
	}, nil
}

func (s *UserServiceImpl) Login(ctx context.Context, request *model.UsersLoginRequest) (*model.UsersLoginResponse, error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	if !strings.Contains(request.Email, "@webmail.uad.ac.id") {
		return nil, helper.SingleError("email", "INVALID_DOMAIN")
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	user, err := s.UserRepository.GetByEmail(tx, request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "Failed to get user")
	}

	if !user.IsVerified {
		return nil, helper.SingleError("user", "NOT_VERIFIED")
	}

	ok, err := helper.CompareEncryptedPassword(s.Viper, user.Password, request.Password)
	if err != nil || !ok {
		return nil, helper.SingleError("credentials", "INVALID")
	}

	accessToken, err := s.Jwt.GenerateAccessToken(user.ID, user.Email, user.Level)
	if err != nil {
		return nil, helper.ServerError(s.Log, "Failed to generate access token")
	}

	refreshToken, err := s.Jwt.GenerateRefreshToken(user.ID, user.Email, user.Level)
	if err != nil {
		return nil, helper.ServerError(s.Log, "Failed to generate refresh token")
	}

	t := time.Now()
	user.LastLogin = t
	if err := s.UserRepository.Update(tx, user); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, helper.ServerError(s.Log, "Failed to commit transaction")
	}

	return &model.UsersLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserServiceImpl) RefreshToken(ctx context.Context, request *model.UsersRefreshTokenRequest) (*model.UserRefreshTokenResponse, error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	claims, err := s.Jwt.ValidateToken(request.RefreshToken)
	if err != nil {
		return nil, helper.SingleError("refresh_token", "INVALID")
	}

	if claims.TokenType != "refresh" {
		return nil, helper.SingleError("refresh_token", "INVALID_TOKEN_TYPE")
	}

	accessToken, err := s.Jwt.GenerateAccessToken(claims.UserID, claims.Email, claims.Level)
	if err != nil {
		return nil, helper.ServerError(s.Log, "Failed to generate access token")
	}

	refreshToken, err := s.Jwt.GenerateRefreshToken(claims.UserID, claims.Email, claims.Level)
	if err != nil {
		return nil, helper.ServerError(s.Log, "Failed to generate refresh token")
	}

	return &model.UserRefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserServiceImpl) VerifyEmail(ctx context.Context, request *model.UsersVerifyEmailRequest) (*model.Response[string], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	user, err := s.UserRepository.GetByVerificationToken(tx, request.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("token", "INVALID")
		}
		return nil, helper.ServerError(s.Log, "Failed to get user")
	}

	if user.IsVerified {
		return nil, helper.SingleError("user", "ALREADY_VERIFIED")
	}

	user.IsVerified = true
	if err := s.UserRepository.Update(tx, user); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, helper.ServerError(s.Log, "Failed to commit transaction")
	}

	message := "Email verified successfully"
	return &model.Response[string]{
		Data: &message,
	}, nil
}

func (s *UserServiceImpl) ResetPassword(ctx context.Context, request *model.UserResetPassword) (*model.Response[string], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	user, err := s.UserRepository.GetByResetPasswordToken(tx, request.Token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("token", "INVALID")
		}
		return nil, helper.ServerError(s.Log, "Failed to get user")
	}

	user.Password = request.Password
	if err := s.UserRepository.Update(tx, user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "Failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, helper.ServerError(s.Log, "Failed to commit transaction")
	}

	message := "Password reset successfully. Please login again."
	return &model.Response[string]{
		Data: &message,
	}, nil
}

func (s *UserServiceImpl) RequestResetPassword(ctx context.Context, request *model.UserResetPasswordRequest) (*model.Response[string], error) {
	if err := s.Validator.Struct(request); err != nil {
		return nil, helper.ValidationError(err)
	}

	tx := s.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	user, err := s.UserRepository.GetByEmail(tx, request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "Failed to get user")
	}

	resetPasswordToken := uuid.NewString()
	user.ResetPasswordToken = resetPasswordToken
	if err := s.UserRepository.Update(tx, user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, helper.SingleError("user", "NOT_FOUND")
		}
		return nil, helper.ServerError(s.Log, "Failed to update user")
	}

	if err := tx.Commit().Error; err != nil {
		return nil, helper.ServerError(s.Log, "Failed to commit transaction")
	}

	// Email template preparation
	var replaceEmail = struct {
		Link string
	}{
		Link: fmt.Sprintf("https://simeru.vercel.app/auth/reset?ref=https%%3A%%2F%%2Fsimeru-scraper.koyeb.app&id=%s", resetPasswordToken),
	}

	tmpl, err := template.ParseFS(templateFS, "template/verify_email.html")
	if err != nil {
		s.Log.Errorf("failed to parse template: %v", err)
		return nil, helper.ServerError(s.Log, "Failed to parse template")
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, &replaceEmail); err != nil {
		s.Log.Errorf("failed to execute template: %v", err)
		return nil, helper.ServerError(s.Log, "Failed to execute template")
	}

	// Send email through message queue
	emailMessage := &messaging.EmailMessage{
		To:      request.Email,
		From:    s.Mail.GetFromEmail(),
		Subject: "[Smrv2] Reset your password",
		Body:    body.String(),
	}

	if err := s.MailProducer.PublishEmailSending(ctx, emailMessage); err != nil {
		return nil, helper.ServerError(s.Log, "Failed to queue email sending for reset password")
	}

	message := "Password reset request sent. Please check your email for the reset link."
	return &model.Response[string]{
		Data: &message,
	}, nil
}
