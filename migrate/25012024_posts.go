package main

import (
	"context"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewCreateTable().Model((*Post)(nil)).Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		_, err := db.NewDropTable().Model((*Post)(nil)).IfExists().Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
