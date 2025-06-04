package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Storage wraps an sqlite database connection.
type Storage struct {
	DB *sql.DB
}

// New opens database by path and initializes schema.
func New(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	s := &Storage{DB: db}
	if err := s.init(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Storage) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS trades(
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            symbol TEXT,
            category TEXT,
            side TEXT,
            qty REAL,
            price REAL,
            ts INTEGER
        );`,
		`CREATE TABLE IF NOT EXISTS assets(
            symbol TEXT PRIMARY KEY,
            qty REAL,
            avg_price REAL
        );`,
	}
	for _, q := range queries {
		if _, err := s.DB.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// SaveTrades stores trades and updates asset quantities and average price.
func (s *Storage) SaveTrades(trades []Trade) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range trades {
		_, err := tx.Exec(`INSERT INTO trades(symbol, category, side, qty, price, ts) VALUES(?,?,?,?,?,?)`,
			t.Symbol, t.Category, t.Side, t.Qty, t.Price, t.Timestamp)
		if err != nil {
			return err
		}
		if t.Side == "Buy" {
			if _, err := tx.Exec(`INSERT INTO assets(symbol, qty, avg_price) VALUES(?,?,?)
                ON CONFLICT(symbol) DO UPDATE SET
                    avg_price=(assets.qty*assets.avg_price + excluded.qty*excluded.avg_price)/(assets.qty+excluded.qty),
                    qty=assets.qty + excluded.qty`, t.Symbol, t.Qty, t.Price); err != nil {
				return err
			}
		} else if t.Side == "Sell" {
			if _, err := tx.Exec(`UPDATE assets SET qty=qty-? WHERE symbol=?`, t.Qty, t.Symbol); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// Asset represents stored asset information.
type Asset struct {
	Symbol   string
	Qty      float64
	AvgPrice float64
}

// Trade represents a trade record.
type Trade struct {
	Symbol    string
	Category  string
	Side      string
	Qty       float64
	Price     float64
	Timestamp int64
}

// Assets returns all assets stored in DB.
func (s *Storage) Assets() ([]Asset, error) {
	rows, err := s.DB.Query(`SELECT symbol, qty, avg_price FROM assets WHERE qty>0`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var as []Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.Symbol, &a.Qty, &a.AvgPrice); err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, rows.Err()
}
