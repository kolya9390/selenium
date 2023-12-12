package modules

import (
	"studentgit.kata.academy/Nikolai/selenium/internal/infrastructure/responder"
	controllerauth "studentgit.kata.academy/Nikolai/selenium/internal/modules/auth/controllerAuth"
	controllergeo "studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/controller_geo"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/repository"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/servis"
	controllerv "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/controller_v"
	repositoryv "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/repository"
	servis_v "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/servis"
)

type Controller struct {
	AuthController	controllerauth.Auther
	GeoController	controllergeo.GeoServiceController
	VacancyController controllerv.VacController
}

func NewControllers(services servis.DadataService, responder responder.Responder, geoRepo repository.GeoRepository,vacRepo repositoryv.VacRepository, vacSevis servis_v.VacancyService) *Controller {

	authController := controllerauth.NewAuth(responder)
	geoController := controllergeo.NewGeoController(services,responder,geoRepo)
	vacContoller := controllerv.NewVacController(vacSevis,responder,vacRepo)

	return &Controller{
		AuthController: authController,
		GeoController: geoController,
		VacancyController: vacContoller,

	}
}
