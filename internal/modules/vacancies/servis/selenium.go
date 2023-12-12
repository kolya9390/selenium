package servis_v

import (
	"fmt"
	"log"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"studentgit.kata.academy/Nikolai/selenium/internal/models"
)

type VacancyService interface {
	ScrapAndSaveVacancies(query string) ([]models.Vacancy, error)
}

type SeleniumServiceImpl struct{}

func NewSeleniumService() *SeleniumServiceImpl {
	return &SeleniumServiceImpl{}
}

const (
	maxTries      = 15
	loadWaitTime  = 5 * time.Second
	baseURL       = "https://career.habr.com/vacancies?page=1&q=%s&type=all"
	titleSelector = "vacancy-card__title"
	descSelector  = "vacancy-card__description"
)

func (s *SeleniumServiceImpl) ScrapAndSaveVacancies(query string) ([]models.Vacancy, error) {
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	chromeCaps := chrome.Capabilities{
		W3C: true,
	}
	caps.AddChrome(chromeCaps)

	var wd selenium.WebDriver
	var err error

	urlPrefix := selenium.DefaultURLPrefix

	// Retry to create WebDriver to handle transient failures
	for i := 1; i <= maxTries; i++ {
		wd, err = selenium.NewRemote(caps, urlPrefix)
		if err != nil {
			log.Println(err)
			if i == maxTries {
				return nil, fmt.Errorf("failed to create WebDriver: %s", err)
			}
			time.Sleep(time.Second) // Wait before retry
			continue
		}
		break
	}
	defer wd.Quit()

	if err := wd.Get(fmt.Sprintf(baseURL, query)); err != nil {
		return nil, fmt.Errorf("failed to load page: %s", err)
	}

	time.Sleep(loadWaitTime) // Wait for the page to load

	titles, err := wd.FindElements(selenium.ByClassName, titleSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find titles: %s", err)
	}

	descriptions, err := wd.FindElements(selenium.ByClassName, descSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to find descriptions: %s", err)
	}

	var vacancies []models.Vacancy

	// Fetch text from elements and create Vacancy objects
	for i := range titles {
		titleText, err := titles[i].Text()
		if err != nil {
			log.Printf("failed to get title text: %s", err)
			continue
		}
		descriptionText, err := descriptions[i].Text()
		if err != nil {
			log.Printf("failed to get description text: %s", err)
			continue
		}

		vacancy := models.Vacancy{
			Title:       titleText,
			Company:     "test",
			Description: descriptionText,
		}
		vacancies = append(vacancies, vacancy)
	}

	// TODO: Add logic to save vacancies to a database or storage

	return vacancies, nil
}
