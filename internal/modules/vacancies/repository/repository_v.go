package repositoryv

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"studentgit.kata.academy/Nikolai/selenium/internal/models"
)

type VacRepository interface {
	SaveVacancy(vacancy models.Vacancy) (int, error) // Вставка в таблицу Вакансий
	GetVacancyByID(vacancyID int) (models.Vacancy, error) // Получаем Вакансию по айди из таблицы вакансий
	DeleteVacancy(vacancyID int) error // Удаляем Вакансию по айди из таблицы вакансий
	SearchInHistory(query string) (bool, error) // Проверка в базе истории поиска а существования такого запроса
	GetVacancyList() ([]models.Vacancy, error) // Выводит Список всех вакансий

	SaveSearchHistory(query string) (int, error) // Вставка строки поиска в Таблицу Serch_hisory
	ListSearchHistory() ([]models.SearchHistory, error) // Вывод список истории поиска
	DeleteSearchHistory(id int) error // Удаление строки поиска по айди

	SaveHistorySearchVacancy(vacancyID, queryID int) error // Вставка айдишников в связующию базу
	GetVacancy(query string) ([]models.Vacancy, error) // Select по двум таблицам и возрат Вакансий
}


type VacRepo struct {
	db *sqlx.DB
}

func NewVacRepository(db *sqlx.DB) *VacRepo {
	return &VacRepo{db: db}
}

func (d *VacRepo) ConnectToDB() error {

	sqlStatementSearch_history := `
CREATE TABLE IF NOT EXISTS search_history (
    id SERIAL PRIMARY KEY,
    query text
);`



	sqlStatementAddress := `
CREATE TABLE IF NOT EXISTS vacancy (
    id SERIAL PRIMARY KEY,
	title VARCHAR(255),
	company VARCHAR(255),
	location VARCHAR(255),
	description TEXT
);`

	sqlStatementHistory_search_address := `
CREATE TABLE IF NOT EXISTS history_search_vacancy (
    id SERIAL PRIMARY KEY,
    search_history_id int,
    vacancy_id int
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



func (r *VacRepo) SaveVacancy(vacancy models.Vacancy) (int, error) {
	query := squirrel.Insert("vacancy").
		Columns("title", "company", "location", "description").
		Values(vacancy.Title, vacancy.Company, vacancy.Location, vacancy.Description).
		Suffix("RETURNING id")

	var id int
	err := query.RunWith(r.db).QueryRow().Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *VacRepo) GetVacancyByID(vacancyID int) (models.Vacancy, error) {
	var vacancy models.Vacancy
	query := squirrel.Select("*").From("vacancy").Where(squirrel.Eq{"id": vacancyID})

	sql, args, err := query.ToSql()
	if err != nil {
		return models.Vacancy{}, err
	}

	err = r.db.Get(&vacancy, sql, args...)
	if err != nil {
		return models.Vacancy{}, err
	}
	return vacancy, nil
}

func (r *VacRepo) DeleteVacancy(vacancyID int) error {
	query := squirrel.Delete("vacancy").Where(squirrel.Eq{"id": vacancyID})

	_, err := query.RunWith(r.db).Exec()
	return err
}

// Проверка в базе истории поиска а существования такого запроса

func (r *VacRepo) SearchInHistory(query string) (bool, error) {
    var exists bool
    // Используем оператор % для поиска похожих запросов в таблице search_history
    err := r.db.QueryRow("SELECT EXISTS (SELECT query FROM search_history WHERE query % $1)", query).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}



func (r *VacRepo) GetVacancyList() ([]models.Vacancy, error) {
	var vacancies []models.Vacancy
	query := squirrel.Select("*").From("vacancy")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v models.Vacancy
		if err := rows.StructScan(&v); err != nil {
			return nil, err
		}
		vacancies = append(vacancies, v)
	}
	return vacancies, nil
}

func (r *VacRepo) SaveSearchHistory(query string) (int, error) {
	queryBuilder := squirrel.Insert("search_history").
		Columns("query").
		Values(query).
		Suffix("RETURNING id")

	var id int
	err := queryBuilder.RunWith(r.db).QueryRow().Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}


func (r *VacRepo) ListSearchHistory() ([]models.SearchHistory, error) {
	var history []models.SearchHistory
	query := squirrel.Select("id", "query", "timestamp").From("search_history")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Queryx(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.SearchHistory
		if err := rows.Scan(&h.ID, &h.Query, &h.Timestamp); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *VacRepo) DeleteSearchHistory(id int) error {
	query := squirrel.Delete("search_history").Where(squirrel.Eq{"id": id})
	_, err := query.RunWith(r.db).Exec()
	return err
}

// Вставка айдишников в связующию базу
func (r *VacRepo) SaveHistorySearchVacancy(vacancyID, queryID int) error {
	// Сохраняем айдишники в связующей базе
	q := squirrel.Insert("history_search_vacancy").Columns("search_history_id", "vacancy_id").
		Values(queryID, vacancyID)

	_, err := q.RunWith(r.db).Exec()
	return err
}

func (r *VacRepo) GetVacancy(query string) ([]models.Vacancy, error) {
    var vacancies []models.Vacancy
    // Выполним запрос, сканируем результаты в структуру Vacancy
    err := r.db.Select(&vacancies, `
        SELECT v.id, v.title, v.company, v.location, v.description
        FROM vacancy v
        JOIN history_search_vacancy hsv ON v.id = hsv.vacancy_id
        JOIN search_history sh ON sh.id = hsv.search_history_id
        WHERE sh.query LIKE $1
    `, "%"+query+"%")
    if err != nil {
        return nil, err
    }
    return vacancies, nil
}
