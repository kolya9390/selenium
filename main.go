package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"studentgit.kata.academy/Nikolai/selenium/internal/config"
	"studentgit.kata.academy/Nikolai/selenium/internal/infrastructure/responder"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/repository"
	"studentgit.kata.academy/Nikolai/selenium/internal/modules/geoservis/servis"
	repositoryv "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/repository"
	servis_v "studentgit.kata.academy/Nikolai/selenium/internal/modules/vacancies/servis"
	"studentgit.kata.academy/Nikolai/selenium/internal/router"
)

func main() {

	env, err := godotenv.Read("app/.env")

	if err != nil {
		log.Println(err)
	}

//	log.Println(env)

	config := &config.AppConf{
		DB: config.DB{
			Host:     env["DB_HOST"],
			Port:     env["DB_PORT"],
			User:     env["DB_USER"],
			Password: env["DB_PASSWORD"],
			Name:     env["DB_NAME"],
		},
		AuthorizationDADATA: config.AuthorizationDADATA{
			ApiKeyValue: env["DADATA_API_KEY"],
			SecretKeyValue: env["DADATA_SECRET_KEY"],
		},

	}

// Инициализация подключения к базе данных
connstr:=  fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
config.DB.Host, config.DB.Port, config.DB.User, config.DB.Password, config.DB.Name)

db, err := sqlx.Open("postgres",connstr)
if err != nil {
	log.Fatalf("Error connecting to the database: %s", err)
}
time.Sleep(time.Second*3)
// Проверка соединения с базой данных
if err := db.Ping(); err != nil {
	log.Fatalf("Error pinging the database: %s", err)
}

defer db.Close()

// Создание экземпляра репозитория с переданным подключением к базе данных
geoRepozitori := repository.NewGeoRepository(db)
vacRepozitori := repositoryv.NewVacRepository(db)

// Create Tabels
/*
err = geoRepozitori.ConnectToDB()

if err!=nil{
	log.Printf("ПРОБЛЕМА conect DB %s",err)
}
*/
err = vacRepozitori.ConnectToDB()

if err!=nil{
	log.Printf("ПРОБЛЕМА conect DB VAC %s",err)
}

vacServis := servis_v.NewSeleniumService()
controllers := modules.NewControllers(&servis.DadataServiceImpl{
	AuthorizationDADATA: config.AuthorizationDADATA},
	responder.NewResponder(zap.New(zap.NewNop().Core())),geoRepozitori,vacRepozitori,vacServis)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router.NewApiRouter(*controllers), // Здесь должен быть ваш обработчик запросов
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Создание канала для получения сигналов остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Println("Starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Ожидание сигнала остановки
	<-stop

	log.Println("Shutting down server...")

	// Создание контекста с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Остановка сервера с использованием graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Println("Server stopped gracefully")


}