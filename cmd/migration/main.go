package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		target     = flag.String("target", "all", "Target datastore (mysql, postgres, redis, elasticsearch, all)")
		command    = flag.String("command", "up", "Migration command (up, down, version)")
		mysqlDSN   = flag.String("mysql-dsn", "root:password@tcp(localhost:3306)/testdb?multiStatements=true", "MySQL DSN")
		postDSN    = flag.String("postgres-dsn", "postgresql://postgres:password@localhost:5432/testdb?sslmode=disable", "PostgreSQL DSN")
		redisDSN   = flag.String("redis-dsn", "redis://localhost:6379/0", "Redis DSN")
		elasticURL = flag.String("elastic-url", "http://localhost:9200", "Elasticsearch URL")
	)
	flag.Parse()

	ctx := context.Background()

	switch strings.ToLower(*target) {
	case "all":
		runMySQLMigration(*mysqlDSN, *command)
		runPostgresMigration(*postDSN, *command)
		runRedisInit(ctx, *redisDSN)
		runElasticsearchInit(ctx, *elasticURL)
	case "mysql":
		runMySQLMigration(*mysqlDSN, *command)
	case "postgres":
		runPostgresMigration(*postDSN, *command)
	case "redis":
		runRedisInit(ctx, *redisDSN)
	case "elasticsearch":
		runElasticsearchInit(ctx, *elasticURL)
	default:
		log.Fatalf("Unknown target: %s", *target)
	}
}

func runMySQLMigration(dsn, command string) {
	m, err := migrate.New(
		"file://test/integration/user/testdata",
		fmt.Sprintf("mysql://%s", dsn),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	executeMigration(m, command)
}

func runPostgresMigration(dsn, command string) {
	m, err := migrate.New(
		"file://test/integration/product/testdata",
		fmt.Sprintf("postgres://%s", strings.TrimPrefix(dsn, "postgresql://")), // Fix DSN format
	)
	if err != nil {
		log.Fatal(err)
	}
	defer m.Close()
	executeMigration(m, command)
}

func runRedisInit(ctx context.Context, dsn string) {
	opts, err := redis.ParseURL(dsn)
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(opts)
	defer rdb.Close()

	// Read and execute Redis commands from init file
	data, err := os.ReadFile("test/integration/cache/testdata/init.redis")
	if err != nil {
		log.Fatal(err)
	}

	commands := strings.Split(string(data), "\n")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}
		args := strings.Fields(cmd)

		// Convert []string to []interface{}
		cmdArgs := make([]interface{}, len(args))
		for i, v := range args {
			cmdArgs[i] = v
		}

		if err := rdb.Do(ctx, cmdArgs...).Err(); err != nil {
			log.Printf("Error executing Redis command %q: %v", cmd, err)
		}
	}
}

func runElasticsearchInit(ctx context.Context, url string) {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{url},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Read template files
	templates, err := filepath.Glob("test/integration/order/testdata/*.json")
	if err != nil {
		log.Fatal(err)
	}

	for _, template := range templates {
		data, err := os.ReadFile(template)
		if err != nil {
			log.Printf("Error reading template %s: %v", template, err)
			continue
		}

		res, err := es.Indices.PutIndexTemplate(
			"test-template",
			strings.NewReader(string(data)),
			es.Indices.PutIndexTemplate.WithContext(ctx),
		)
		if err != nil {
			log.Printf("Error applying template %s: %v", template, err)
			continue
		}
		defer res.Body.Close()
	}
}

// validateMigration ตรวจสอบสถานะของ migration และรายงานปัญหาที่พบ
func validateMigration(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("error checking migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d", version)
	}

	return nil
}

func executeMigration(m *migrate.Migrate, command string) {
	var err error

	// ตรวจสอบสถานะก่อนทำ migration
	if err := validateMigration(m); err != nil {
		if command != "force" {
			log.Printf("Warning: %v", err)
			log.Println("You may need to use 'force' command to fix this")
		}
	}

	switch command {
	case "up":
		err = m.Up()
	case "down":
		err = m.Down()
	case "version":
		version, dirty, verErr := m.Version()
		if verErr != nil {
			err = verErr
		} else {
			log.Printf("Current version: %d (dirty: %v)", version, dirty)
			return
		}
	case "force":
		version, dirty, verErr := m.Version()
		if verErr != nil {
			err = verErr
		} else {
			targetVersion := int(version)
			if dirty {
				targetVersion-- // ถ้า dirty ให้ถอยกลับ 1 version
			}
			log.Printf("Forcing migration to version %d", targetVersion)
			err = m.Force(targetVersion)
		}
	default:
		log.Fatalf("Unknown command: %s", command)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// ตรวจสอบสถานะหลังทำ migration
	if err := validateMigration(m); err != nil {
		log.Printf("Warning: Migration might not be in a clean state: %v", err)
	}
}
