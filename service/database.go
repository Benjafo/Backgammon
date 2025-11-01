package service

import (
	"log"
	"net/http"

	"backgammon/repository"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
		db := repository.GetDB()

		err := db.Ping(r.Context())
		if err != nil {
			log.Println("Database ping failed:", err)
			http.Error(w, "Database ping failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Database connection successful!"))
}
