package entity
import (
	"time"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	UserID       uuid.UUID          `bson:"user_id" json:"userId"` // Penerima Notifikasi [cite: 284]
	Title        string             `bson:"title" json:"title"` 
	Message      string             `bson:"message" json:"message"` 
	Type         string             `bson:"type" json:"type"` // offer, order_status, new_order 
	RelatedID    uuid.UUID          `bson:"related_id" json:"relatedId"` // ID Order/Offer 
	IsRead       bool               `bson:"is_read" json:"isRead"` 
	CreatedAt    time.Time          `bson:"created_at" json:"createdAt"` 
}
