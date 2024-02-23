package main

import (
	"context"
	"errors"
	"flag"

	"cloud.google.com/go/spanner"
	"github.com/rs/zerolog/log"
	"github.com/yukia3e/go-spanner-selectall-test/internal/util"
	"google.golang.org/api/iterator"
)

type Singer struct {
	ID   int64  `spanner:"ArtistId"`
	Name string `spanner:"Name"`
}

type SingerLowCase struct {
	ID   int64  `spanner:"artistid"`
	Name string `spanner:"name"`
}

func main() {
	ctx := context.Background()

	withLenient := flag.Bool("l", false, "Use lenient mode")
	flag.Parse()

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

	if _, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		_, err := txn.Update(ctx, spanner.Statement{
			SQL: "INSERT INTO Singers (ArtistId, Name) VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');",
		})
		return err
	}); err != nil {
		log.Fatal().Err(err).Msg("Failed to insert data")
	}

	// Case 1: row.ToStruct => Success!
	log.Info().Msg("start row.ToStruct")
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
	log.Info().Msg("end row.ToStruct")

	// Case 2: lower-case spanner.SelectAll WithLenient => Success!
	log.Info().Msg("start lower-case spanner.SelectAll WithLenient")
	func() {
		iter := client.Single().Read(ctx, "Singers", spanner.AllKeys(), []string{"ArtistId", "Name"})
		defer iter.Stop()

		var singers []*SingerLowCase
		if err := spanner.SelectAll(iter, &singers, spanner.WithLenient()); err != nil {
			// The error will not be output because it will panic in SelectAll
			log.Fatal().Err(err).Msg("Failed to select all")
		}
		log.Info().Interface("singers", singers).Msg("lower-case spanner.SelectAll WithLenient => Success!")
	}()
	log.Info().Msg("end lower-case spanner.SelectAll WithLenient")

	// Case 3: lower-case spanner.SelectAll => Success!
	log.Info().Msg("start lower-case spanner.SelectAll")
	func() {
		iter := client.Single().Read(ctx, "Singers", spanner.AllKeys(), []string{"ArtistId", "Name"})
		defer iter.Stop()

		var singers []*SingerLowCase
		if err := spanner.SelectAll(iter, &singers); err != nil {
			// The error will not be output because it will panic in SelectAll
			log.Fatal().Err(err).Msg("Failed to select all")
		}
		log.Info().Interface("singers", singers).Msg("lower-case spanner.SelectAll => Success!")
	}()
	log.Info().Msg("end lower-case spanner.SelectAll")

	// memo: Switch between the case with and without WithLenient flag by flag to end with panic
	if withLenient != nil && *withLenient {
		// Case 4: spanner.SelectAll WithLenient => Failed!
		log.Info().Msg("start spanner.SelectAll WithLenient")
		func() {
			iter := client.Single().Read(ctx, "Singers", spanner.AllKeys(), []string{"ArtistId", "Name"})
			defer iter.Stop()

			var singers []*Singer
			if err := spanner.SelectAll(iter, &singers, spanner.WithLenient()); err != nil {
				// The error will not be output because it will panic in SelectAll
				log.Fatal().Err(err).Msg("Failed to select all")
			}
			// Expected result, but the following line will not be output because it will panic in SelectAll
			log.Info().Interface("singers", singers).Msg("spanner.SelectAll WithLenient => Success!")
		}()
		log.Info().Msg("end spanner.SelectAll WithLenient")
	} else {
		// Case 5: spanner.SelectAll => Failed!
		log.Info().Msg("start spanner.SelectAll")
		func() {
			iter := client.Single().Read(ctx, "Singers", spanner.AllKeys(), []string{"ArtistId", "Name"})
			defer iter.Stop()

			var singers []*Singer
			if err := spanner.SelectAll(iter, &singers); err != nil {
				// The error will not be output because it will panic in SelectAll
				log.Fatal().Err(err).Msg("Failed to select all")
			}
			// Expected result, but the following line will not be output because it will panic in SelectAll
			log.Info().Interface("singers", singers).Msg("spanner.SelectAll => Success!")
		}()
		log.Info().Msg("end spanner.SelectAll")
	}
}
