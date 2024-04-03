package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Ad struct {
	gorm.Model
	Title      string      `json:"title"`
	StartAt    time.Time   `json:"startAt"`
	EndAt      time.Time   `json:"endAt"`
	Conditions []Condition `gorm:"foreignKey:AdID" json:"conditions"`
}

type Condition struct {
	gorm.Model
	AdID     uint           // Foreign key for the Ad
	AgeStart int            `gorm:"type:integer" json:"ageStart"`
	AgeEnd   int            `gorm:"type:integer" json:"ageEnd"`
	Gender   pq.StringArray `gorm:"type:text[]" json:"gender"`
	Country  pq.StringArray `gorm:"type:text[]" json:"country"`
	Platform pq.StringArray `gorm:"type:text[]" json:"platform"`
}

type ResponseItem struct {
	Title string    `json:"title"`
	EndAt time.Time `json:"endAt"`
}

type Response struct {
	Result  string          `json:"result"`
	Message string          `json:"message,omitempty"`
	Items   *[]ResponseItem `json:"items,omitempty"`
}

var db *gorm.DB

func ConnectDatabase() *gorm.DB {
	username := viper.GetString("db.username")
	password := viper.GetString("db.password")
	host := viper.GetString("db.host")
	port := viper.GetString("db.port")
	dbname := viper.GetString("db.name")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Taipei", host, username, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	err = db.AutoMigrate(&Ad{}, &Condition{})
	if err != nil {
		panic("failed to migrate the database: " + err.Error())
	}
	return db
}

func CreateAdHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	var ad Ad
	json.Unmarshal(body, &ad)

	result := db.Create(&ad)
	if result.Error != nil {
		log.Printf("Failed to insert ad \"%s\" into the database: %s\n", ad.Title, result.Error.Error())
		json.NewEncoder(w).Encode(Response{
			Result:  "failed",
			Message: result.Error.Error(),
		})
	} else {
		log.Printf("Ad \"%s\" inserted successfully with ID: %d\n", ad.Title, ad.ID)
		json.NewEncoder(w).Encode(Response{
			Result: "success",
		})
	}
}

func ListAdHandler(w http.ResponseWriter, r *http.Request) {
	parm := r.URL.Query()
	offset := 0
	limit := 5
	var err error
	if parm.Has("offset") {
		offset, err = strconv.Atoi(parm["offset"][0])
		if err != nil {
			offset = 0
		}
	}
	if parm.Has("limit") {
		limit, _ = strconv.Atoi(parm["limit"][0]) // Atoi gives 0 if error, no need to handle error
	}
	if limit < 1 || limit > 100 {
		json.NewEncoder(w).Encode(Response{
			Result:  "failed",
			Message: "Range of limit must between 1 and 100",
		})
		return
	}

	currentTime := time.Now()

	query := db.Distinct("ads.*").
		Model(&Ad{}).
		Joins("JOIN conditions ON conditions.ad_id = ads.id").
		Where("ads.start_at <= ? AND ? <= ads.end_at", currentTime, currentTime).
		Offset(offset).
		Limit(limit).
		Order("end_at ASC")

	if parm.Has("age") {
		query = query.Where("(conditions.age_start = 0 AND conditions.age_end = 0) OR (conditions.age_start <= ? AND ? <= conditions.age_end)", parm["age"], parm["age"])
	} else {
		query = query.Where("conditions.age_start = 0 AND conditions.age_end = 0")
	}

	if parm.Has("gender") {
		query = query.Where("array_length(conditions.gender, 1) IS NULL OR ? = ANY(conditions.gender)", parm["gender"])
	} else {
		query = query.Where("array_length(conditions.gender, 1) IS NULL")
	}

	if parm.Has("country") {
		query = query.Where("array_length(conditions.country, 1) IS NULL OR ? = ANY(conditions.country)", parm["country"])
	} else {
		query = query.Where("array_length(conditions.country, 1) IS NULL")
	}

	if parm.Has("platform") {
		query = query.Where("array_length(conditions.platform, 1) IS NULL OR ? = ANY(conditions.platform)", parm["platform"])
	} else {
		query = query.Where("array_length(conditions.platform, 1) IS NULL")
	}

	var ads []ResponseItem
	err = query.
		Select("title", "end_at").
		Find(&ads).Error

	if err != nil {
		json.NewEncoder(w).Encode(Response{
			Result:  "failed",
			Message: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(Response{
		Result: "success",
		Items:  &ads,
	})
}

func main() {
	viper.SetConfigFile("config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic("Error reading config file")
	}

	db = ConnectDatabase()

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/ad", CreateAdHandler).Methods("POST")
	r.HandleFunc("/api/v1/ad", ListAdHandler).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server on :8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
