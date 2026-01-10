package postgres

import "github.com/jmoiron/sqlx"

type Adapter struct {
	DB *sqlx.DB
}

func NewPostgresAdapter(dsn string) (*Adapter, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &Adapter{DB: db}, nil
}

func (p *Adapter) GetDB() *sqlx.DB {
	return p.DB
}

func (p *Adapter) HealthCheck() error {
	return p.DB.Ping()
}

func (p *Adapter) Close() error {
	return p.DB.Close()
}

func (p *Adapter) RunQuery(query string, args ...any) (*sqlx.Rows, error) {
	return p.DB.Queryx(query, args...)
}

func (p *Adapter) DescribeSchema() (string, error) {
	rows, err := p.DB.Queryx(
		` SELECT table_name, column_name, data_type 
 				FROM information_schema.columns 
 				WHERE table_schema = 'public' 
 				ORDER BY table_name, ordinal_position`)
	if err != nil {
		return "", err
	}

	defer func(rows *sqlx.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	schema := ""

	for rows.Next() {
		var table, column, dtype string
		if err := rows.Scan(&table, &column, &dtype); err != nil {
			return "", err
		}
	}

	return schema, nil
}
