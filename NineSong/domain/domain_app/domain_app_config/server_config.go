package domain_app_config

import (
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/usecase"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type AppServerConfig struct {
	ID          primitive.ObjectID `bson:"_id"`
	ServerName  string             `bson:"serverName" bson:"server_name"`
	URL         string             `bson:"url" bson:"url"`
	UserName    string             `bson:"userName" bson:"user_name"`
	Password    string             `bson:"password" bson:"password"`
	LastLoginAt time.Time          `bson:"lastLoginAt" bson:"last_login_at,omitempty"`
	Type        string             `bson:"type" bson:"type"`
}

// AppServerConfigUsecase defines the usecase interface for app server configuration.
// It embeds the generic BaseUsecase to provide standard CRUD operations.
type AppServerConfigUsecase interface {
	usecase.BaseUsecase[AppServerConfig]
}
