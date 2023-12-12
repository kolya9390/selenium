package controllerv

import (
	"encoding/json"
	"log"
	"net/http"

	"studentgit.kata.academy/Nikolai/selenium/internal/infrastructure/responder"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/repository"
	servis_v "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/servis"
)

type VacController interface {
	Search(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

type VacancyController struct {
	VacancyService servis_v.VacancyService
	VacRepo repositoryv.VacRepository
	responder.Responder
}

func NewVacController(servis servis_v.VacancyService ,responder responder.Responder, vacRep repositoryv.VacRepository) *VacancyController {

	return &VacancyController{VacancyService: servis, Responder: responder,VacRepo: vacRep}
}

func (vc *VacancyController) Search(w http.ResponseWriter, r *http.Request) {

	var requestBody RequestSerch
	//var resp ResponseSerch

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		vc.Responder.ErrorBadRequest(w, err)
		log.Println("Decoder Body")
		return
    }

	log.Println(requestBody)
	// Проверка, есть ли такой запрос в истории
	exists, err := vc.VacRepo.SearchInHistory(requestBody.Query)
	if err != nil {
		log.Printf("FindAddressByQueryAndHistory %s", err)
	}

	// Если запрос не найден в истории, выполняем скрапинг
	if !exists {
		respVacancy ,err := vc.VacancyService.ScrapAndSaveVacancies(requestBody.Query)
		if err != nil {
			vc.Responder.ErrorInternal(w,err)
			log.Println("RespData")
			log.Println(err)
		return
		}

		// Сохраняем запрос в истории
		queryID, err := vc.VacRepo.SaveSearchHistory(requestBody.Query)
		if err != nil {
			log.Printf("Failed to save search history: %v\n", err)
			return
		}

		for _, v := range respVacancy {
			vacancyID, err := vc.VacRepo.SaveVacancy(v)
			if err != nil {
				log.Printf("SaveVacancy: %v\n", err)
				return
			}

			// Сохраняем идентификаторы в связующую базу
			err = vc.VacRepo.SaveHistorySearchVacancy(vacancyID, queryID)
			if err != nil {
				log.Printf("Failed to save search-vacancy association: %v\n", err)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		vc.OutputJSON(w, ResponseVacancy{Vacancyes: respVacancy})
		return
	}

	// Получаем вакансии из базы данных
	vacancies, err := vc.VacRepo.GetVacancy(requestBody.Query)
	if err != nil {
		log.Printf("GetVacancy: %v\n", err)
		return
	}

	// Отправляем вакансии клиенту
	w.WriteHeader(http.StatusOK)
	vc.OutputJSON(w, ResponseVacancy{Vacancyes: vacancies})
}



func (vc *VacancyController) Delete(w http.ResponseWriter, r *http.Request) {
	// Обработка запроса на удаление вакансии
}

func (vc *VacancyController) GetByID(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestGetByID
	//var resp ResponseSerch

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		vc.Responder.ErrorBadRequest(w, err)
		log.Println("Decoder Body")
		return
    }

	vacancies, err := vc.VacRepo.GetVacancyByID(requestBody.ID)
	if err != nil {
		log.Printf("GetVacancy: %v\n", err)
		return
	}

	// Отправляем вакансии клиенту
	w.WriteHeader(http.StatusOK)
	vc.OutputJSON(w, ResponseGetByID{ID: requestBody.ID,Vacancye: vacancies})

}

func (vc *VacancyController) List(w http.ResponseWriter, r *http.Request) {
	// Обработка запроса на удаление вакансии по иайди из базы
}
