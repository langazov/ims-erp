package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	log, _ := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "test",
	})

	cfg := config.MongoDBConfig{
		URI:      "mongodb://localhost:27017",
		Database: "erp_system",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongodb, err := repository.NewMongoDB(cfg, log)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer mongodb.Close(ctx)

	userStore := repository.NewReadModelStore(mongodb, "users", log)

	result, err := userStore.FindOne(ctx, map[string]interface{}{
		"email":    "admin@erp.local",
		"tenantid": "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		fmt.Printf("FindOne error: %v\n", err)
		return
	}
	if result == nil {
		fmt.Println("No user found")
		return
	}

	fmt.Printf("Type of result: %T\n", result)

	// Print as JSON to see structure
	if jsonBytes, err := json.MarshalIndent(result, "", "  "); err == nil {
		fmt.Printf("Result as JSON:\n%s\n", string(jsonBytes))
	}

	// Try bson.M
	if m, ok := result.(bson.M); ok {
		fmt.Println("Successfully converted to bson.M")
		fmt.Printf("email field: %v\n", m["email"])
	} else {
		fmt.Println("Failed to convert to bson.M")
	}
}
