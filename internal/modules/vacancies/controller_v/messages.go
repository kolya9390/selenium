package controllerv

import "studentgit.kata.academy/Nikolai/selenium/internal/models"

//go:generate easytags $GOFILE

type RequestSerch struct {
//	ID    int    `json:"id"`
	Query string `json:"query"`
}

type ResponseSerch struct {
//	VacancyID   int    `json:"vacancy_id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type ResponseVacancy struct {
	Vacancyes []models.Vacancy `json:"vacancyes"`
}

type RequestGetByID struct {
	ID int `json:"id"`
}

type ResponseGetByID struct {
	ID       int            `json:"id"`
	Vacancye models.Vacancy `json:"vacancye"`
}
