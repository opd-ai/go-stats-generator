package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Mongo provides MongoDB-backed storage for analysis results.
type Mongo struct {
	client     *mongo.Client
	collection *mongo.Collection
	mu         sync.RWMutex
}

// mongoDocument represents the MongoDB document structure.
type mongoDocument struct {
	ID     string          `bson:"_id"`
	Status string          `bson:"status"`
	Report json.RawMessage `bson:"report,omitempty"`
	Error  *string         `bson:"error,omitempty"`
}

// NewMongo creates a new MongoDB storage instance.
// connectionString format: "mongodb://host:port" or "mongodb+srv://user:password@cluster/database"
func NewMongo(connectionString string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database("go_stats_generator")
	collection := database.Collection("analysis_results")

	m := &Mongo{
		client:     client,
		collection: collection,
	}

	if err := m.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return m, nil
}

// initSchema creates indexes for the collection.
func (m *Mongo) initSchema() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "_id", Value: 1}},
	}

	_, err := m.collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

// Store saves an analysis result.
func (m *Mongo) Store(result *AnalysisResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var reportJSON json.RawMessage
	var errorText *string

	if result.Report != nil {
		reportJSON, _ = json.Marshal(result.Report)
	}
	if result.Error != nil {
		errStr := result.Error.Error()
		errorText = &errStr
	}

	doc := mongoDocument{
		ID:     result.ID,
		Status: result.Status,
		Report: reportJSON,
		Error:  errorText,
	}

	opts := options.Replace().SetUpsert(true)
	m.collection.ReplaceOne(ctx, bson.M{"_id": result.ID}, doc, opts)
}

// Get retrieves an analysis result by ID.
func (m *Mongo) Get(id string) (*AnalysisResult, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var doc mongoDocument
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		return nil, false
	}

	return convertDocToResult(&doc), true
}

// List returns all stored analysis results.
func (m *Mongo) List() []*AnalysisResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return []*AnalysisResult{}
	}
	defer cursor.Close(ctx)

	var results []*AnalysisResult
	for cursor.Next(ctx) {
		var doc mongoDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		results = append(results, convertDocToResult(&doc))
	}

	return results
}

// convertDocToResult converts a MongoDB document to an AnalysisResult.
func convertDocToResult(doc *mongoDocument) *AnalysisResult {
	result := &AnalysisResult{
		ID:     doc.ID,
		Status: doc.Status,
	}

	if len(doc.Report) > 0 {
		var report metrics.Report
		if err := json.Unmarshal(doc.Report, &report); err == nil {
			result.Report = &report
		}
	}

	if doc.Error != nil {
		result.Error = fmt.Errorf("%s", *doc.Error)
	}

	return result
}

// Delete removes an analysis result by ID.
func (m *Mongo) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err == nil && result.DeletedCount > 0
}

// Clear removes all stored analysis results.
func (m *Mongo) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	m.collection.DeleteMany(ctx, bson.M{})
}

// Close closes the MongoDB connection.
func (m *Mongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.client.Disconnect(ctx)
}
