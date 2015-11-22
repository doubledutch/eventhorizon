package eventhorizon

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	// PostgreSQL driver
	_ "github.com/lib/pq"
)

// ErrCouldNotCreateTables returned when necessary tables could not be created.
var ErrCouldNotCreateTables = errors.New("could not create tables")

func initDB(conn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, ErrCouldNotDialDB
	}

	db.MapperFunc(strings.ToLower)

	return db, nil
}

// PostgresReadRepository implements an Postgres repository of read models.
type PostgresReadRepository struct {
	db      *sqlx.DB
	table   string
	factory func() interface{}
	stmts   map[string]string
}

// NewPostgresReadRepository creates a new PostgresReadRepository.
func NewPostgresReadRepository(conn, table string) (*PostgresReadRepository, error) {
	db, err := initDB(conn)
	if err != nil {
		return nil, err
	}

	create := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  data jsonb
);
`, table)

	_, err = db.Exec(create)
	if err != nil {
		fmt.Println(err)
		return nil, ErrCouldNotCreateTables
	}

	stmts := map[string]string{
		"save":    fmt.Sprintf("INSERT INTO %s (data) VALUES ($1)", table),
		"find":    fmt.Sprintf("SELECT * FROM %s WHERE data->>'id'=$1", table),
		"findall": fmt.Sprintf("SELECT * FROM %s", table),
		"remove":  fmt.Sprintf("DELETE FROM %s WHERE data->>'id'=$1", table),
		"clear":   fmt.Sprintf("DELETE FROM %s", table),
		"update":  fmt.Sprintf("UPDATE %s set data=$1 WHERE data->>'id'=$2", table),
	}

	return &PostgresReadRepository{
		db:    db,
		table: table,
		stmts: stmts,
	}, nil
}

// Save saves a read model with id to the repository.
func (r *PostgresReadRepository) Save(id UUID, model interface{}) error {
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}

	existing, err := r.Find(id)
	if err != nil && err != ErrModelNotFound {
		return err
	}

	if existing != nil {
		// Update
		_, err = r.db.Exec(r.stmts["update"], b, id.String())
		if err != nil {
			fmt.Println(err)
			return ErrCouldNotSaveModel
		}
		return nil
	}

	// Insert
	_, err = r.db.Exec(r.stmts["save"], b)
	if err != nil {
		return ErrCouldNotSaveModel
	}

	return nil
}

type postgresModel struct {
	Data []byte
}

// Find returns one read model with using an id.
func (r *PostgresReadRepository) Find(id UUID) (interface{}, error) {
	if r.factory == nil {
		return nil, ErrModelNotSet
	}

	var pModel postgresModel
	err := r.db.Get(&pModel, r.stmts["find"], id.String())
	if err != nil && err == sql.ErrNoRows {
		return nil, ErrModelNotFound
	} else if err != nil {
		return nil, err
	}

	model := r.factory()
	err = json.Unmarshal(pModel.Data, model)
	return model, err
}

// FindAll returns all read models in the repository.
func (r *PostgresReadRepository) FindAll() ([]interface{}, error) {
	if r.factory == nil {
		return nil, ErrModelNotSet
	}

	var pModels []postgresModel
	err := r.db.Select(&pModels, r.stmts["findall"])
	if err != nil {
		return nil, err
	}

	models := make([]interface{}, len(pModels))
	for i, pModel := range pModels {
		model := r.factory()
		err = json.Unmarshal(pModel.Data, model)
		if err != nil {
			return nil, err
		}
		models[i] = model
	}
	return models, nil
}

// Remove removes a read model with id from the repository.
func (r *PostgresReadRepository) Remove(id UUID) error {
	result, err := r.db.Exec(r.stmts["remove"], id.String())
	if num, err := result.RowsAffected(); err != nil {
		return err
	} else if num == 0 {
		return ErrModelNotFound
	}
	return err
}

// SetModel sets a factory function that creates concrete model types.
func (r *PostgresReadRepository) SetModel(factory func() interface{}) {
	r.factory = factory
}

// Clear clears the read model table.
func (r *PostgresReadRepository) Clear() error {
	_, err := r.db.Exec(r.stmts["clear"])
	return err
}

// Close closes the postgres db connection.
func (r *PostgresReadRepository) Close() error {
	return r.db.Close()
}
