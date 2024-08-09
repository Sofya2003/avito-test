package postgres

import (
	"database/sql"
	"time"

	"avtest/internal/store"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db := &PostgresDB{DB: dbConn}
	err = db.CreateTable()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *PostgresDB) CreateTable() error {
	_, err := db.DB.Exec(`
-- 		DROP TABLE flats;
-- 		DROP TABLE houses;
-- 		DROP TABLE users;
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			type TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS houses (
			id SERIAL PRIMARY KEY,
			house_number INTEGER UNIQUE NOT NULL,
			address TEXT NOT NULL,
			year_built INTEGER NOT NULL,
			developer TEXT,
			created_at TIMESTAMP NOT NULL,
			last_flat_added_at TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS flats (
			id SERIAL PRIMARY KEY,
			house_id INTEGER REFERENCES houses(house_number) ON DELETE CASCADE,
			flat_number INTEGER NOT NULL,
			price INTEGER NOT NULL,
			rooms INTEGER NOT NULL,
			status TEXT NOT NULL,
		    moderator TEXT DEFAULT '' NOT NULL
		);
	`)
	return err
}

// User methods
func (db *PostgresDB) CreateUser(user *store.User) error {
	_, err := db.DB.Exec("INSERT INTO users (email, password, type) VALUES ($1, $2, $3)",
		user.Email, user.Password, user.Type)
	return err
}

func (db *PostgresDB) GetUserByEmail(email string) (*store.User, error) {
	row := db.DB.QueryRow("SELECT id, email, password, type FROM users WHERE email = $1", email)
	var user store.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// House methods
func (db *PostgresDB) CreateHouse(house *store.House) error {
	_, err := db.DB.Exec(`
		INSERT INTO houses (house_number, address, year_built, developer, created_at, last_flat_added_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		house.HouseNumber, house.Address, house.YearBuilt, house.Developer, house.CreatedAt, house.LastFlatAddedAt)
	return err
}

func (db *PostgresDB) GetHouseByID(id int64) (*store.House, error) {
	row := db.DB.QueryRow(`
		SELECT house_number, address, year_built, developer, created_at, last_flat_added_at
		FROM houses WHERE id = $1`, id)
	var house store.House
	err := row.Scan(&house.HouseNumber, &house.Address, &house.YearBuilt, &house.Developer,
		&house.CreatedAt, &house.LastFlatAddedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &house, nil
}

func (db *PostgresDB) UpdateHouse(house *store.House) error {
	_, err := db.DB.Exec(`
		UPDATE houses SET address = $1, year_built = $2, developer = $3,
		created_at = $4, last_flat_added_at = $5 WHERE id = $6`,
		house.Address, house.YearBuilt, house.Developer, house.CreatedAt,
		house.LastFlatAddedAt, house.HouseNumber)
	return err
}

// Flat methods
func (db *PostgresDB) CreateFlat(flat *store.Flat) error {
	_, err := db.DB.Exec(`INSERT INTO flats (house_id, flat_number, price, rooms, status) 
		VALUES ($1, $2, $3, $4, $5)`,
		flat.HouseNumber, flat.FlatNumber, flat.Price, flat.Rooms, flat.Status)
	return err
}

func (db *PostgresDB) GetFlat(houseNumber, flatNumber int64) (*store.Flat, error) {
	var flat store.Flat
	row := db.DB.QueryRow(`
		SELECT id, house_id, flat_number, price, rooms, status 
		FROM flats 
		WHERE house_id = $1 AND flat_number = $2`, houseNumber, flatNumber)
	err := row.Scan(&flat.ID, &flat.HouseNumber, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &flat, nil
}

func (db *PostgresDB) UpdateHouseFlatTime(time time.Time) error {
	_, err := db.DB.Exec(`UPDATE houses SET last_flat_added_at = $1`, time)
	return err
}

func (db *PostgresDB) GetFlatsByHouseID(houseID int64, userType string) ([]store.Flat, error) {
	var flats []store.Flat
	query := `SELECT id, house_id, flat_number, price, rooms, status FROM flats WHERE house_id = $1`
	args := []interface{}{houseID}

	if userType == "client" {
		query += ` AND status = $2`
		args = append(args, "approved")
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var flat store.Flat
		if err := rows.Scan(&flat.ID, &flat.HouseNumber, &flat.FlatNumber, &flat.Price, &flat.Rooms, &flat.Status); err != nil {
			return nil, err
		}
		flats = append(flats, flat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return flats, nil
}

func (db *PostgresDB) GetFlatStatus(houseID int64, flatNumber int64) (store.Flat, error) {
	var f store.Flat
	row := db.DB.QueryRow(`SELECT * FROM flats WHERE house_id = $1 AND flat_number = $2`,
		houseID, flatNumber)

	if err := row.Scan(&f.ID, &f.HouseNumber, &f.FlatNumber, &f.Price, &f.Rooms, &f.Status, &f.Moderator); err != nil {
		return store.Flat{}, err
	}

	return f, nil
}

func (db *PostgresDB) UpdateFlat(flat *store.Flat, token string) error {
	_, err := db.DB.Exec(`
		UPDATE flats SET status = $1, moderator = $2
		WHERE flat_number = $3 AND house_id = $4`,
		flat.Status, token, flat.FlatNumber, flat.HouseNumber)
	return err
}
