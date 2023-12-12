package repository

import (

	"github.com/jmoiron/sqlx"
)

type GeoRepository interface {
	InsertSearchHistory(query string) (int, error) // Вставка строки поиска в Таблицу search_history
	InsertAddress(region, geoLat, geoLon string) (int, error) // Вставка адреса, и координат в Таблицу address
	InsertHistorySearchAddress(searchHistoryID, addressID int) error // Вставка айд в Таблицу history_search_address
	SearchInHistory(query string) (bool, error)
	FindAddressByQueryAndHistory(query string) ([]AddressData, error)
}

type geoRepository struct {
	db *sqlx.DB
}

func NewGeoRepository(db *sqlx.DB) *geoRepository {
	return &geoRepository{db: db}
}

func (d *geoRepository) ConnectToDB() error {


	sqlStatementSearch_history := `
CREATE TABLE IF NOT EXISTS search_history (
    id SERIAL PRIMARY KEY,
    query text
);`

	sqlStatementAddress := `
CREATE TABLE IF NOT EXISTS address (
    id SERIAL PRIMARY KEY,
    region text,
    geo_lat text,
    geo_lon text
);`

	sqlStatementHistory_search_address := `
CREATE TABLE IF NOT EXISTS history_search_address (
    id SERIAL PRIMARY KEY,
    search_history_id int,
    address_id int
);`

setTrgm := `CREATE EXTENSION IF NOT EXISTS pg_trgm;`


	_, err := d.db.Exec(sqlStatementSearch_history)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(sqlStatementAddress)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(sqlStatementHistory_search_address)
	if err != nil {
		return err
	}


	_, err = d.db.Exec(setTrgm)
	if err != nil {
		return err
	}


	return nil

}



func (gr *geoRepository) InsertSearchHistory(query string) (int, error) {
	var id int
	err := gr.db.QueryRowx("INSERT INTO search_history (query) VALUES ($1) RETURNING id", query).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}


func (gr *geoRepository) InsertAddress(region, geoLat, geoLon string) (int, error) {
	var id int
	err := gr.db.QueryRowx("INSERT INTO address (region, geo_lat, geo_lon) VALUES ($1, $2, $3) RETURNING id", region, geoLat, geoLon).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}


func (gr *geoRepository) InsertHistorySearchAddress(searchHistoryID, addressID int) error {
	_, err := gr.db.Exec("INSERT INTO history_search_address (search_history_id, address_id) VALUES ($1, $2)", searchHistoryID, addressID)
	if err != nil {
		return err
	}
	return nil
}


type AddressData struct {
    Region string
    GeoLat string
    GeoLon string
}


func (gr *geoRepository) FindAddressByQueryAndHistory(query string) ([]AddressData, error) {
    var addresses []AddressData
    // выполните запрос и сканируйте результат в структуру AddressData
    err := gr.db.Select(&addresses, `
        SELECT a.geo_lat as GeoLat, a.geo_lon as GeoLon, a.region as Region
        FROM address a
        JOIN history_search_address hsa ON a.id = hsa.address_id
        JOIN search_history sh ON sh.id = hsa.search_history_id
        WHERE sh.query LIKE $1
    `, "%"+query+"%")
    if err != nil {
        return nil, err
    }
    return addresses, nil
}





func (gr *geoRepository) SearchInHistory(query string) (bool, error) {
    var exists bool
    // Используем оператор % для поиска похожих запросов в таблице search_history
    err := gr.db.QueryRow("SELECT EXISTS (SELECT query FROM search_history WHERE query % $1)", query).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}
