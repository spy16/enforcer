package postgres

import (
	"context"
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

// Store implements the enforcer storage using PostgresQL.
type Store struct{ db *sql.DB }

func (st *Store) withReadTx(ctx context.Context, fns ...txnFn) error {
	return st.withTx(ctx, &sql.TxOptions{ReadOnly: true}, fns)
}

func (st *Store) withReadWriteTx(ctx context.Context, fns ...txnFn) error {
	return st.withTx(ctx, &sql.TxOptions{ReadOnly: false}, fns)
}

func (st *Store) withTx(ctx context.Context, opts *sql.TxOptions, fns []txnFn) error {
	tx, err := st.db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	for _, fn := range fns {
		if fnErr := fn(tx); fnErr != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("rollback failed as well. unpredictable")
			}
			return fnErr
		}
	}

	return tx.Commit()
}

type txnFn func(tx *sql.Tx) error

const schema = `
CREATE TABLE IF NOT EXISTS campaigns (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS enrolments (
	id SERIAL PRIMARY KEY,
	campaign_id INTEGER NOT NULL,
	actor_id TEXT NOT NULL,
	started_at TIMESTAMP NOT NULL,
	ends_at TIMESTAMP NOT NULL,
	completed_steps JSON NOT NULL,
	FOREIGN KEY (campaign_id) REFERENCES campaigns (id),
	UNIQUE(campaign_id, actor_id)
);

CREATE TABLE IF NOT EXISTS tags (
	id SERIAL PRIMARY KEY,
	tag TEXT NOT NULL,
	UNIQUE(tag)
);

CREATE TABLE IF NOT EXISTS campaigns_tags (
	campaign_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	PRIMARY KEY(campaign_id, tag_id),
	FOREIGN KEY (tag_id) REFERENCES tags (id),
	FOREIGN KEY (campaign_id) REFERENCES campaigns (id)
);
`
