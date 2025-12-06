package mongodb

import (
	"context"
	"fmt"
	"time"
	entity "home-market/internal/domain" // Asumsi entity diimpor

	"go.mongodb.org/mongo-driver/mongo"
)

// PLACEHOLDERS
const (
	DatabaseName     = "random_home_market"
	CollectionStatus = "history_status"
	CollectionNotifs = "notifications" // <--- TAMBAHKAN DEFINISI COLLECTION BARU
)

type LogRepository interface {
	SaveHistoryStatus(doc *entity.HistoryStatus) error
	SaveNotification(doc *entity.Notification) error
}

type logRepository struct {
    // FIX 1: Ubah field dari collection ke client
	client *mongo.Client 
}

// NewLogRepository: Constructor untuk inisialisasi LogRepository
func NewLogRepository(client *mongo.Client) LogRepository {
	// FIX 2: Simpan objek Client
	return &logRepository{
		client: client,
	}
}

// FR-ORDER-02/03: Simpan Riwayat Status (Disesuaikan)
func (r *logRepository) SaveHistoryStatus(doc *entity.HistoryStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// FIX 3: Akses collection secara dinamis dari r.client
	collection := r.client.Database(DatabaseName).Collection(CollectionStatus)

	_, err := collection.InsertOne(ctx, doc)

	if err != nil {
		return fmt.Errorf("failed to insert history status to Mongo: %w", err)
	}
	return nil
}

// Implementasi method baru
func (r *logRepository) SaveNotification(doc *entity.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// FIX 4: Akses collection 'notifications' secara dinamis dari r.client
	collection := r.client.Database(DatabaseName).Collection(CollectionNotifs)

	_, err := collection.InsertOne(ctx, doc)

	if err != nil {
		return fmt.Errorf("failed to insert notification to Mongo: %w", err)
	}

	return nil
}