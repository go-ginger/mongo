package mongo

import (
	"github.com/go-ginger/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BaseModel struct {
	models.BaseModel `bson:"-" json:"-"`

	ID        *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time           `bson:"created_at,omitempty" json:"created_at,omitempty" mongo:"insert_only"`
	UpdatedAt time.Time           `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
	DeletedAt *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func (base *BaseModel) updateFromBase() {
	base.CreatedAt = base.BaseModel.CreatedAt
	base.UpdatedAt = base.BaseModel.UpdatedAt
	base.DeletedAt = base.BaseModel.DeletedAt
}

func (base *BaseModel) HandleCreateDefaultValues() {
	base.BaseModel.HandleCreateDefaultValues()
	base.updateFromBase()
}

func (base *BaseModel) HandleUpdateDefaultValues() {
	base.BaseModel.HandleUpdateDefaultValues()
	base.updateFromBase()
}

func (base *BaseModel) HandleUpsertDefaultValues() {
	base.CreatedAt = time.Now().UTC()
	base.UpdatedAt = time.Now().UTC()
}

func (base *BaseModel) HandleDeleteDefaultValues() {
	base.BaseModel.HandleDeleteDefaultValues()
	base.updateFromBase()
}

func (base *BaseModel) GetID() interface{} {
	return base.ID
}

func (base *BaseModel) GetIDString() string {
	return base.ID.Hex()
}

func (base *BaseModel) SetID(id interface{}) {
	oid := id.(primitive.ObjectID)
	base.ID = &oid
}
