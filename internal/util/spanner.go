package util

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"github.com/rs/zerolog/log"
)

const (
	defaultSpannerEmulatorHost = "localhost:9010"
)

func CreateTestDatabase(ctx context.Context, extraStatements []string) (*spanner.Client, func()) {
	project := "test-project"
	instance := "test-instance"
	dbName := "test-database"

	if err := createSpannerEmulatorDatabase(ctx, project, instance, dbName, extraStatements); err != nil {
		log.Fatal().Err(err).Msg("Failed to create test database")
	}

	client, err := spanner.NewClient(ctx, spannerDatabaseName(project, instance, dbName))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Spanner client")
	}
	deferFn := func() {
		client.Close()
		if err := dropSpannerEmulatorDatabase(ctx, project, instance, dbName); err != nil {
			log.Fatal().Err(err).Msg("Failed to drop test database")
		}
	}

	return client, deferFn
}

func createSpannerEmulatorDatabase(ctx context.Context, project, instance, dstDB string, extraStatements []string) error {
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", defaultSpannerEmulatorHost)
	}

	client, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Drop the database if it already exists
	if err := client.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
		Database: spannerDatabaseName(project, instance, dstDB),
	}); err != nil {
		return err
	}

	// Create the new database
	op, err := client.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          spannerInstanceName(project, instance),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", dstDB),
		ExtraStatements: extraStatements,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func dropSpannerEmulatorDatabase(ctx context.Context, project, instance, db string) error {
	if os.Getenv("SPANNER_EMULATOR_HOST") == "" {
		os.Setenv("SPANNER_EMULATOR_HOST", defaultSpannerEmulatorHost)
	}

	client, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	return client.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
		Database: spannerDatabaseName(project, instance, db),
	})
}

func spannerInstanceName(project, instance string) string {
	return fmt.Sprintf("projects/%s/instances/%s", project, instance)
}

func spannerDatabaseName(project, instance, database string) string {
	return fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, instance, database)
}
