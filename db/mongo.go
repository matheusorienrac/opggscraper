package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/matheusorienrac/opggscraper/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName         = "live-lol-esports-stats"
	collectionName = "ranked_stats"
	connectTimeout = 10 * time.Second
)

// Client wraps the MongoDB client
type Client struct {
	client *mongo.Client
}

// ConnectDB establishes a connection to the MongoDB instance.
func ConnectDB(uri string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the primary
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB!")
	return &Client{client: client}, nil
}

// Disconnect closes the MongoDB connection.
func (c *Client) Disconnect(ctx context.Context) {
	if c.client != nil {
		if err := c.client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}

// SaveChampionStats saves or updates champion stats in the database.
// It uses Upsert based on ChampionName, Patch, and Tier.
func (c *Client) SaveChampionStats(ctx context.Context, stats model.RankedChampionStats) error {
	collection := c.client.Database(dbName).Collection(collectionName)

	filter := bson.M{
		"championName": stats.ChampionName,
		"patch":        stats.Patch,
		"tier":         stats.Tier,
	}

	update := bson.M{
		"$set": stats, // Update all fields including matchups and scrapedAt
	}

	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to upsert champion stats for %s (Patch: %s, Tier: %s): %w", stats.ChampionName, stats.Patch, stats.Tier, err)
	}

	log.Printf("Successfully saved/updated stats for %s (Patch: %s, Tier: %s)", stats.ChampionName, stats.Patch, stats.Tier)
	return nil
}
