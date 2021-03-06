package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/k8shop/go-rest-api/pkg/handlers"
	"github.com/k8shop/go-rest-api/pkg/handlers/middleware"
	"github.com/k8shop/go-rest-api/pkg/informer"
	"github.com/k8shop/go-rest-api/pkg/models"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
)

func main() {
	db, err := initDB()
	defer db.Close()
	informer := informer.NewInformer([]string{os.Getenv("KAFKA_HOST") + ":" + os.Getenv("KAFKA_PORT")})
	defer informer.Close()

	if err != nil {
		panic(err)
	}
	router := mux.NewRouter()
	router.Use(middleware.AddCommonHeaders)
	err = handlers.Register(router, db, informer)
	if err != nil {
		panic(err)
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(os.Getenv("CORS_ORIGINS"), ";"),
		AllowCredentials: true,
		AllowedMethods:   strings.Split(os.Getenv("CORS_METHODS"), ";"),
	})

	handler := c.Handler(router)

	http.ListenAndServe(":8080", handler)
}

func initDB() (*gorm.DB, error) {
	connectionURL := fmt.Sprintf(
		"%v:%v@tcp(%v)/?parseTime=true",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASS"),
		os.Getenv("DATABASE_HOST"),
	)

	log.Printf("Connecting to DB: %s", connectionURL)
	preDB, err := sql.Open("mysql", connectionURL)
	if err != nil {
		return nil, err
	}

	_, err = preDB.Exec("CREATE DATABASE IF NOT EXISTS " + os.Getenv("DATABASE_NAME"))
	if err != nil {
		return nil, err
	}

	preDB.Close()

	connectionURL = fmt.Sprintf(
		"%v:%v@tcp(%v)/%v?parseTime=true",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASS"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_NAME"),
	)
	log.Printf("Connecting to DB: %s", connectionURL)
	db, err := gorm.Open("mysql", connectionURL)
	if err != nil {
		return nil, err
	}

	db.Debug().AutoMigrate(&models.Product{}, &models.User{}, &models.VerificationCode{})

	return db, nil
}
