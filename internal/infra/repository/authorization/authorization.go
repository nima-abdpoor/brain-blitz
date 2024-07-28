package repository

import (
	entity "BrainBlitz.com/game/entity/auth"
	"BrainBlitz.com/game/internal/core/port/repository"
	model "BrainBlitz.com/game/internal/infra/repository/mongo"
	"BrainBlitz.com/game/pkg/errmsg"
	"BrainBlitz.com/game/pkg/richerror"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Authorization struct {
	DB *mongo.Database
}

func NewAuthorizationRepo(db *mongo.Database) repository.AuthorizationRepository {
	return Authorization{
		DB: db,
	}
}

func (a Authorization) GetUserPermissionTitles(role entity.Role) ([]entity.PermissionTitle, error) {
	const op = "repository.GetUserPermissionTitles"

	var accessControl model.AccessControl
	filter := bson.D{{"role_type", role.String()}}
	//todo bind context.Background() to service not hardCoding!
	err := a.DB.Collection("access_control").FindOne(context.Background(), filter).Decode(&accessControl)
	if err != nil {
		log.Println(err)
		if err == mongo.ErrNoDocuments {
			return []entity.PermissionTitle{}, nil
		}
		return []entity.PermissionTitle{}, richerror.
			New(op).
			WithKind(richerror.KindUnexpected).
			WithMessage(errmsg.SomeThingWentWrong)
	}

	var permissionTitles []entity.PermissionTitle
	for _, title := range accessControl.Permissions {
		permissionTitles = append(permissionTitles, entity.PermissionTitle(title))
	}
	return permissionTitles, nil
}
