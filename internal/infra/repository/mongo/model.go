package mongo

import "go.mongodb.org/mongo-driver/bson/primitive"

type AccessControl struct {
	ID          primitive.ObjectID `bson:"_id"`
	RoleType    string             `bson:"role_type,omitempty"`
	Permissions []string           `bson:"permissions,omitempty"`
}
