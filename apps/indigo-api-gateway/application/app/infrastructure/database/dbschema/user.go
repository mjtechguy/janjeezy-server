package dbschema

import (
	"menlo.ai/indigo-api-gateway/app/domain/user"
	"menlo.ai/indigo-api-gateway/app/infrastructure/database"
)

func init() {
	database.RegisterSchemaForAutoMigrate(User{})
}

type User struct {
	BaseModel
	Name          string `gorm:"type:varchar(100);not null"`
	Email         string `gorm:"type:varchar(255);uniqueIndex;not null"`
	PublicID      string `gorm:"type:varchar(50);uniqueIndex;not null"`
	Enabled       bool
	Organizations []OrganizationMember `gorm:"foreignKey:UserID"`
	Projects      []ProjectMember      `gorm:"foreignKey:UserID"`
	IsGuest       bool
	PasswordHash  *string `gorm:"type:text"`
}

func NewSchemaUser(u *user.User) *User {
	var passwordHash *string
	if u.PasswordHash != "" {
		passwordHash = &u.PasswordHash
	}
	return &User{
		BaseModel: BaseModel{
			ID: u.ID,
		},
		Name:         u.Name,
		Email:        u.Email,
		Enabled:      u.Enabled,
		PublicID:     u.PublicID,
		IsGuest:      u.IsGuest,
		PasswordHash: passwordHash,
	}
}

func (u *User) EtoD() *user.User {
	passwordHash := ""
	if u.PasswordHash != nil {
		passwordHash = *u.PasswordHash
	}
	return &user.User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		Enabled:      u.Enabled,
		PublicID:     u.PublicID,
		CreatedAt:    u.CreatedAt,
		IsGuest:      u.IsGuest,
		PasswordHash: passwordHash,
	}
}
