package mongo

import (
	"github.com/go-ginger/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type BaseModel struct {
	models.BaseModel `bson:"-" json:"-"`

	ID        interface{} `bson:"_id" json:"id,omitempty"`
	CreatedAt time.Time   `bson:"created_at" json:"created_at,omitempty"`
	UpdatedAt time.Time   `bson:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time  `bson:"deleted_at" json:"deleted_at,omitempty"`
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

func (base *BaseModel) HandleDeleteDefaultValues() {
	base.BaseModel.HandleDeleteDefaultValues()
	base.updateFromBase()
}

func (base *BaseModel) SetID(id interface{}) {
	base.ID = id.(primitive.ObjectID)
}
