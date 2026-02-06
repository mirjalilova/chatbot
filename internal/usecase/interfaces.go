// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"time"

	"chatbot/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test

type (
	// UserRepo -.
	UserRepoI interface {
		// Login(ctx context.Context, req *entity.LoginReq) (*entity.LoginRes, error)
		CheckExist(ctx context.Context, phone string) (bool, error)
		Create(ctx context.Context, req *entity.CreateUser) (*entity.UserInfo, error)
		CreateGuest(ctx context.Context,ip, ua string) (string, error)
		GetGuestByIPAndUA(ctx context.Context, ip, ua string) (string, error)
		GetById(ctx context.Context, req *entity.ById) (*entity.UserInfo, error)
		GetAll(ctx context.Context, req *entity.Filter, status string) (*entity.UserList, error)
		Update(ctx context.Context, req *entity.UpdateUser) error
		Delete(ctx context.Context, req *entity.ById) error
		GetByPhone(ctx context.Context, phone string) (*entity.GetByPhone, error)
		GetMe(ctx context.Context, id string) (*entity.GetMe, error)

		GetByEmail(ctx context.Context, email string) (*entity.UserInfo, error)
		CreateGoogleUser(ctx context.Context, u *entity.CreateGoogleUser) (string, error)
	}

	// RestrictionRepo -.
	RestrictionRepoI interface {
		GetById(ctx context.Context, req *entity.ById) (*entity.Restriction, error)
		GetAll(ctx context.Context, req *entity.Filter) (*entity.ListRestriction, error)
		Update(ctx context.Context, req *entity.UpdateRestriction) error
		// Delete(ctx context.Context, req *entity.ById) error
	}

	// ChatRepo -.
	ChatRepoI interface {
		Create(ctx context.Context, req *entity.ChatCreate) error
		CreateChatRoom(ctx context.Context, req *entity.ChatRoomCreate) (string, error)
		GetChatRoomByUserId(ctx context.Context, id *entity.GetChatRoomReq) (*entity.ChatRoomList, error)
		GetChatRoomChat(ctx context.Context, id *entity.ById, limit, offset int) (*entity.ChatList, error)
		Check(ctx context.Context, user_id, chatRoomID string) (int, error) 
		DeleteChatRoom(ctx context.Context, id *entity.ById) error
	}

	// DashboardRepo -.
	DashboardRepoI interface {
		GetUserAndRequestCount(ctx context.Context, fromDate, toDate time.Time) (*[]entity.DashboardActiveUsers, error)
	}

	// PDFRepo -.
	PDFRepoI interface {
		Create(ctx context.Context, req *entity.CreatePdfCategory) error
		Update(ctx context.Context, req *entity.UpdatePdfCategory) error
		CreateItem(ctx context.Context, req *entity.CretatePdfCategoryItem) error
		UpdateItem(ctx context.Context, req *entity.UpdatePdfCategoryItem) error
		ListPDFCategoryItem(ctx context.Context, id string) (*[]entity.ListPdfCategoryItem, error)
		DeleteCategory(ctx context.Context, req string) error
		DeleteCategoryItem(ctx context.Context, req string) error
		ListPDFCategory(ctx context.Context) (*[]entity.ListPdfCategoryItem, error)
	}
)
