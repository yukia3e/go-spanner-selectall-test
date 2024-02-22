package main

import (
	"context"
	"errors"

	"cloud.google.com/go/spanner"
	"github.com/rs/zerolog/log"
	"github.com/yukia3e/go-spanner-selectall-test/internal/util"
	"google.golang.org/api/iterator"
)

type Singer struct {
	ID   int64  `spanner:"ArtistId"`
	Name string `spanner:"Name"`
}

func main() {
	ctx := context.Background()

	// Create a test database
	client, deferFn := util.CreateTestDatabase(
		ctx,
		[]string{
			`CREATE TABLE Singers (
				ArtistId INT64 NOT NULL,
				Name STRING(1024),
			) PRIMARY KEY (ArtistId)`,
		},
	)
	defer deferFn()

	// Insert test data
	if _, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		_, err := txn.Update(ctx, spanner.Statement{
			SQL: "INSERT INTO Singers (ArtistId, Name) VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');",
		})
		return err
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to insert data")
	}

	// row.ToStruct => Success!
	func() {
		rows := client.Single().Read(ctx, "Singers", spanner.Key{1}, []string{"ArtistId", "Name"})
		defer rows.Stop()
		var singers []*Singer
		for {
			row, err := rows.Next()
			if err != nil {
				if errors.Is(err, iterator.Done) {
					break
				}

				log.Fatal().Err(err).Msg("Failed to read row")
			}

			var singer Singer
			if err := row.ToStruct(&singer); err != nil {
				log.Fatal().Err(err).Msg("Failed to convert row to struct")
			}
			singers = append(singers, &singer)
		}
		log.Info().Interface("singers", singers).Msg("row.ToStruct => Success!")
	}()

	// spanner.SelectAll => Failed!
	func() {
		iter := client.Single().Read(ctx, "Singers", spanner.AllKeys(), []string{"ArtistId", "Name"})
		defer iter.Stop()

		var singers []*Singer
		if err := spanner.SelectAll(iter, &singers, spanner.WithLenient()); err != nil {
			// The error will not be output because it will panic in SelectAll
			log.Fatal().Err(err).Msg("Failed to select all")
		}
		// Expected result, but the following line will not be output because it will panic in SelectAll
		log.Info().Interface("singers", singers).Msg("spanner.SelectAll => Success!")
	}()
}
