package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/ims-erp/system/internal/config"
	"github.com/ims-erp/system/internal/repository"
	"github.com/ims-erp/system/pkg/logger"
)

const (
	DefaultTenantID       = "00000000-0000-0000-0000-000000000001"
	DefaultAdminEmail     = "admin@erp.local"
	DefaultAdminPassword  = "Admin123!"
	DefaultAdminFirstName = "System"
	DefaultAdminLastName  = "Administrator"
)

func main() {
	mongoURI := flag.String("mongodb", "mongodb://localhost:27017", "MongoDB connection URI")
	database := flag.String("database", "erp_system", "MongoDB database name")
	flag.Parse()

	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		*mongoURI = uri
	}

	log, err := logger.New(logger.Config{
		Level:       "info",
		Format:      "json",
		ServiceName: "seed-admin",
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Sync()

	cfg := config.MongoDBConfig{
		URI:      *mongoURI,
		Database: *database,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongodb, err := repository.NewMongoDB(cfg, log)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close(ctx)

	userStore := repository.NewReadModelStore(mongodb, "users", log)

	existingUser, err := userStore.FindOne(ctx, map[string]interface{}{
		"email":    DefaultAdminEmail,
		"tenantid": DefaultTenantID,
	})
	if err != nil {
		log.Fatalf("Failed to check for existing user: %v", err)
	}

	if existingUser != nil {
		log.Info("Admin user already exists", "email", DefaultAdminEmail)
		updateExistingAdmin(ctx, userStore, existingUser)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	now := time.Now().UTC()
	adminUser := map[string]interface{}{
		"_id":           uuid.New().String(),
		"id":            uuid.New().String(),
		"tenantid":      DefaultTenantID,
		"email":         DefaultAdminEmail,
		"passwordhash":  string(hash),
		"firstname":     DefaultAdminFirstName,
		"lastname":      DefaultAdminLastName,
		"phone":         "",
		"role":          "admin",
		"status":        "active",
		"tenantrole":    "admin",
		"permissions":   []string{"*"},
		"mfaenabled":    false,
		"mfasecret":     "",
		"loginattempts": 0,
		"createdat":     now,
		"updatedat":     now,
	}

	if err := userStore.Save(ctx, adminUser); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Info("Default admin user created successfully",
		"email", DefaultAdminEmail,
		"tenantId", DefaultTenantID,
	)
	fmt.Printf(`
==================================================
Default Admin Account Created
==================================================
Email:    %s
Password: %s
TenantID: %s
Role:     admin
==================================================
`, DefaultAdminEmail, DefaultAdminPassword, DefaultTenantID)
}

func updateExistingAdmin(ctx context.Context, userStore *repository.ReadModelStore, existingUser interface{}) {
	userMap, ok := existingUser.(map[string]interface{})
	if !ok {
		log.Println("WARNING: Could not parse existing user, skipping update")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("WARNING: Failed to hash new password:", err)
		return
	}

	userMap["passwordhash"] = string(hash)
	userMap["role"] = "admin"
	userMap["tenantrole"] = "admin"
	userMap["permissions"] = []string{"*"}
	userMap["status"] = "active"
	userMap["updatedat"] = time.Now().UTC()

	err = userStore.Update(ctx, map[string]interface{}{"_id": userMap["_id"]}, map[string]interface{}{
		"$set": userMap,
	})
	if err != nil {
		log.Println("WARNING: Failed to update admin user:", err)
		return
	}

	log.Println("INFO: Admin user updated successfully", "email", DefaultAdminEmail)
	fmt.Printf(`
==================================================
Default Admin Account Updated
==================================================
Email:    %s
Password: %s
==================================================
`, DefaultAdminEmail, DefaultAdminPassword)
}
