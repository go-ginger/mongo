package mongo

import (
	"github.com/go-ginger/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BaseModel struct {
	models.BaseModel `bson:"-" json:"-"`

	ID        *primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty" dl:"read_only"`
	CreatedAt time.Time           `bson:"created_at,omitempty" json:"created_at,omitempty" mongo:"insert_only" dl:"read_only"`
	UpdatedAt *time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty" dl:"read_only"`
	DeletedAt *time.Time          `bson:"deleted_at,omitempty" json:"deleted_at,omitempty" dl:"read_only"`
}

func (base *BaseModel) updateFromBase() {
	if !base.BaseModel.CreatedAt.IsZero() {
		base.CreatedAt = base.BaseModel.CreatedAt
	}
	if base.BaseModel.UpdatedAt != nil {
		base.UpdatedAt = base.BaseModel.UpdatedAt
	}
	if base.BaseModel.DeletedAt != nil {
		base.DeletedAt = base.BaseModel.DeletedAt
	}
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
	now := time.Now().UTC()
	base.CreatedAt = time.Now().UTC()
	base.UpdatedAt = &now
}

func (base *BaseModel) HandleDeleteDefaultValues() {
	base.BaseModel.HandleDeleteDefaultValues()
	base.updateFromBase()
}

func (base *BaseModel) GetID() interface{} {
	return base.ID
}

func (base *BaseModel) GetIDString() string {
	if base.ID != nil {
		return base.ID.Hex()
	}
	return ""
}

func (base *BaseModel) SetID(id interface{}) {
	oid := id.(primitive.ObjectID)
	base.ID = &oid
}
