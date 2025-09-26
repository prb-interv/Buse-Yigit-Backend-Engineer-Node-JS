package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite" 
	"goweather/pkg/types"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(dbPath string) (*Database, error) {
	
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			return nil, fmt.Errorf("database file creation failed: %v", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %v", err)
	}

	// test
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection test failed: %v", err)
	}

	database := &Database{db: db}
	
	if err := database.createTable(); err != nil {
		return nil, fmt.Errorf("table creation failed: %v", err)
	}

	log.Printf("✅ SQLite database connection successful: %s", dbPath)
	return database, nil
}

func (d *Database) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS weather_queries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		location TEXT NOT NULL,
		service_1_temperature REAL NOT NULL,
		service_2_temperature REAL NOT NULL,
		request_count INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("table creation failed: %v", err)
	}

	log.Println("✅ weather_queries table created")
	return nil
}


func (d *Database) SaveWeatherQuery(query *types.WeatherQuery) error {
	stmt := `
	INSERT INTO weather_queries (location, service_1_temperature, service_2_temperature, request_count)
	VALUES (?, ?, ?, ?)`

	result, err := d.db.Exec(stmt, query.Location, query.Service1Temp, query.Service2Temp, query.RequestCount)
	if err != nil {
		return fmt.Errorf("data save failed: %v", err)
	}

	// get ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ID get failed: %v", err)
	}

	query.ID = int(id)
	log.Printf("✅ Weather query saved: ID=%d, Location=%s", query.ID, query.Location)
	return nil
}

func (d *Database) GetWeatherQueries() ([]types.WeatherQuery, error) {
	query := `
	SELECT id, location, service_1_temperature, service_2_temperature, request_count
	FROM weather_queries
	ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("data get failed: %v", err)
	}
	defer rows.Close()

	var queries []types.WeatherQuery
	for rows.Next() {
		var q types.WeatherQuery
		err := rows.Scan(&q.ID, &q.Location, &q.Service1Temp, &q.Service2Temp, &q.RequestCount)
		if err != nil {
			return nil, fmt.Errorf("data read failed: %v", err)
		}
		queries = append(queries, q)
	}

	return queries, nil
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}
