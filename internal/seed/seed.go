package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	truncate := flag.Bool("truncate", false, "Truncate all tables before seeding")
	seedOnly := flag.Bool("seed", false, "Seed the database (default true if no truncate)")
	flag.Parse()

	// Load config
	cfg, err := cmd.LoadConfig()
	if err != nil {
		fmt.Printf("[FATAL] Could not load EnvConfig: %v\n", err)
		os.Exit(1)
	}
	cmd.Env = cfg

	// Check Environment
	if cfg.Environment != "DEV" {
		fmt.Println("❌ Seeding/Truncating is only allowed in DEV environment.")
		os.Exit(1)
	}

	// Initialize Logger
	log, err := logger.InitLogger(cfg.Environment)
	if err != nil {
		fmt.Printf("[FATAL]: Could not initialize Logger: %v\n", err)
		os.Exit(1)
	}
	logger.Log = log

	// Initialize DB Pool
	if err := cmd.InitDBPool(); err != nil {
		logger.Log.Error("[FATAL]: Could not initialize DB Pool: ", err)
		os.Exit(1)
	}
	defer cmd.DBPool.Close()

	ctx := context.Background()
	seeder := NewSeeder()

	if *truncate {
		if err := seeder.TruncateDB(ctx); err != nil {
			logger.Log.Error("[FATAL]: Truncate failed: ", err)
			os.Exit(1)
		}
	}

	// If truncate is false, we default to seed. If truncate is true, we only seed if seedOnly is true or it's implicit?
	// Let's make it simple: if --truncate is passed, it truncates. If --seed is passed (or nothing), it seeds.
	// Actually, usually one might want to truncate AND then seed.

	if *seedOnly || !*truncate {
		if err := seeder.SeedDB(ctx); err != nil {
			logger.Log.Error("[FATAL]: Seeding failed: ", err)
			os.Exit(1)
		}
	}
}

type Seeder struct {
	queries *db.Queries
}

func NewSeeder() *Seeder {
	return &Seeder{
		queries: db.New(),
	}
}

func (s *Seeder) TruncateDB(ctx context.Context) error {
	fmt.Println("🗑️ Truncating tables...")
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := s.queries.TruncateTables(ctx, conn); err != nil {
		return err
	}
	fmt.Println("✔ Tables truncated successfully!")
	return nil
}

func (s *Seeder) SeedDB(ctx context.Context) error {
	fmt.Println("✔ Starting database seeding...")

	// 1. Seed Users
	users, err := s.seedUsers(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}
	fmt.Printf("✔ Seeded %d users\n", len(users))

	// 2. Seed Hubs
	hubs, err := s.seedHubs(ctx)
	if err != nil {
		return fmt.Errorf("failed to seed hubs: %w", err)
	}
	fmt.Printf("✔ Seeded %d hubs\n", len(hubs))

	// 3. Seed Items
	items, err := s.seedItems(ctx, users, hubs)
	if err != nil {
		return fmt.Errorf("failed to seed items: %w", err)
	}
	fmt.Printf("✔ Seeded %d items\n", len(items))

	// 4. Seed Claims
	claims, err := s.seedClaims(ctx, items, users)
	if err != nil {
		return fmt.Errorf("failed to seed claims: %w", err)
	}
	fmt.Printf("✔ Seeded %d claims\n", len(claims))

	// 5. Seed Audit Logs
	if err := s.seedAuditLogs(ctx, users, items, hubs, claims); err != nil {
		return fmt.Errorf("failed to seed audit logs: %w", err)
	}
	fmt.Println("✔ Seeded audit logs")

	fmt.Println("✔ Database seeding completed successfully!")
	return nil
}

func (s *Seeder) seedUsers(ctx context.Context) ([]uuid.UUID, error) {
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var userIDs []uuid.UUID
	password, _ := pkg.Hash("password123")

	users := []struct {
		Name  string
		Email string
		Role  db.UserRole
	}{
		{"Admin User", "admin@nyx.com", db.UserRoleADMIN},
		{"John Doe", "john@example.com", db.UserRoleUSER},
		{"Jane Smith", "jane@example.com", db.UserRoleUSER},
		{"Alice Johnson", "alice@example.com", db.UserRoleUSER},
		{"Bob Williams", "bob@example.com", db.UserRoleUSER},
		{"Charlie Brown", "charlie@example.com", db.UserRoleUSER},
		{"David Miller", "david@example.com", db.UserRoleUSER},
		{"Eva Davis", "eva@example.com", db.UserRoleUSER},
		{"Frank Wilson", "frank@example.com", db.UserRoleUSER},
		{"Grace Taylor", "grace@example.com", db.UserRoleUSER},
	}

	for _, u := range users {
		trustScore := int32(80 + rand.Intn(21)) // 80-100
		id, err := s.queries.SeedUser(ctx, conn, db.SeedUserParams{
			Name:     u.Name,
			Email:    u.Email,
			Password: password,
			Role:     u.Role,
			TrustScore: pgtype.Int4{
				Int32: trustScore,
				Valid: true,
			},
		})
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}

func (s *Seeder) seedHubs(ctx context.Context) ([]uuid.UUID, error) {
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var hubIDs []uuid.UUID
	hubs := []struct {
		Name      string
		Address   string
		Lat, Long string
	}{
		{"Central Station", "123 Main St", "40.7128", "-74.0060"},
		{"North Campus Hub", "456 North Ave", "40.7580", "-73.9855"},
		{"Library Drop-off", "789 Knowledge Way", "40.7829", "-73.9654"},
		{"Student Center", "321 Student Ln", "40.7295", "-73.9965"},
		{"Gymnasium Lost & Found", "654 Fitness Blvd", "40.7484", "-73.9857"},
	}

	for _, h := range hubs {
		id, err := s.queries.SeedHub(ctx, conn, db.SeedHubParams{
			Name:      h.Name,
			Address:   pgtype.Text{String: h.Address, Valid: true},
			Latitude:  pgtype.Text{String: h.Lat, Valid: true},
			Longitude: pgtype.Text{String: h.Long, Valid: true},
			Contact:   pgtype.Text{String: "contact@hub.com", Valid: true},
		})
		if err != nil {
			return nil, err
		}
		hubIDs = append(hubIDs, id)
	}
	return hubIDs, nil
}

func (s *Seeder) seedItems(ctx context.Context, users []uuid.UUID, hubs []uuid.UUID) ([]uuid.UUID, error) {
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var itemIDs []uuid.UUID
	types := []db.ItemType{db.ItemTypeLOST, db.ItemTypeFOUND}
	statuses := []db.ItemStatus{db.ItemStatusOPEN, db.ItemStatusPENDINGCLAIM, db.ItemStatusRESOLVED, db.ItemStatusARCHIVED}

	for i := range 30 {
		userID := users[rand.Intn(len(users))]
		hubID := hubs[rand.Intn(len(hubs))]
		itemType := types[rand.Intn(len(types))]
		status := statuses[rand.Intn(len(statuses))]

		var currentHubID uuid.NullUUID
		if itemType == db.ItemTypeFOUND {
			currentHubID = uuid.NullUUID{UUID: hubID, Valid: true}
		} else {
			currentHubID = uuid.NullUUID{Valid: false}
		}

		name := fmt.Sprintf("Item %d", i+1)
		desc := fmt.Sprintf("Description for item %d. Looks like a generic object.", i+1)
		locDesc := "Near the entrance"
		lat := fmt.Sprintf("40.%d", 7000+rand.Intn(1000))
		long := fmt.Sprintf("-73.%d", 9000+rand.Intn(1000))
		timeAt := time.Now().Add(-time.Duration(rand.Intn(100)) * time.Hour)

		id, err := s.queries.SeedItem(ctx, conn, db.SeedItemParams{
			UserID:              userID,
			HubID:               currentHubID,
			Name:                name,
			Description:         pgtype.Text{String: desc, Valid: true},
			Type:                itemType,
			Status:              status,
			LocationDescription: pgtype.Text{String: locDesc, Valid: true},
			Latitude:            pgtype.Text{String: lat, Valid: true},
			Longitude:           pgtype.Text{String: long, Valid: true},
			TimeAt:              pgtype.Timestamptz{Time: timeAt, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		itemIDs = append(itemIDs, id)
	}
	return itemIDs, nil
}

func (s *Seeder) seedClaims(ctx context.Context, items []uuid.UUID, users []uuid.UUID) ([]uuid.UUID, error) {
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var claimIDs []uuid.UUID
	statuses := []db.ClaimStatus{db.ClaimStatusPENDING, db.ClaimStatusAPPROVED, db.ClaimStatusREJECTED}

	for i := 0; i < 15; i++ {
		itemID := items[rand.Intn(len(items))]
		claimantID := users[rand.Intn(len(users))]
		status := statuses[rand.Intn(len(statuses))]

		proof := "This is mine, I lost it yesterday."
		score := 0.5 + rand.Float64()*0.5 // 0.5 - 1.0

		id, err := s.queries.SeedClaim(ctx, conn, db.SeedClaimParams{
			ItemID:          itemID,
			ClaimantID:      claimantID,
			Status:          status,
			ProofText:       pgtype.Text{String: proof, Valid: true},
			SimilarityScore: pgtype.Float8{Float64: score, Valid: true},
		})

		if err != nil {
			continue
		}
		claimIDs = append(claimIDs, id)
	}
	return claimIDs, nil
}

func (s *Seeder) seedAuditLogs(ctx context.Context, users, items, hubs, claims []uuid.UUID) error {
	conn, err := cmd.DBPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	actions := []string{"CREATED", "UPDATED", "DELETED", "VIEWED"}
	type target struct {
		Type db.TargetType
		IDs  []uuid.UUID
	}
	targets := []target{
		{db.TargetTypeITEM, items},
		{db.TargetTypeUSER, users},
		{db.TargetTypeHUB, hubs},
		{db.TargetTypeCLAIM, claims},
	}

	for i := 0; i < 50; i++ {
		actorID := users[rand.Intn(len(users))]
		action := actions[rand.Intn(len(actions))]
		t := targets[rand.Intn(len(targets))]

		if len(t.IDs) == 0 {
			continue
		}
		targetID := t.IDs[rand.Intn(len(t.IDs))]

		err := s.queries.SeedAuditLog(ctx, conn, db.SeedAuditLogParams{
			ActorID:    uuid.NullUUID{UUID: actorID, Valid: true},
			Action:     action,
			TargetType: t.Type,
			TargetID:   uuid.NullUUID{UUID: targetID, Valid: true},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
