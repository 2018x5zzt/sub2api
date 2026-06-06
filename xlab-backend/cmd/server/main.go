package main

import (
	"log"
	"net/http"

	"github.com/2018x5zzt/xlab-backend/internal/config"
	"github.com/2018x5zzt/xlab-backend/internal/core"
	"github.com/2018x5zzt/xlab-backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	client := core.NewClient(cfg.CoreAPIBaseURL, cfg.CoreTimeout)
	router := httpapi.NewRouter(client)

	log.Printf("xlab backend listening on %s, core=%s", cfg.ServerAddr, cfg.CoreAPIBaseURL)
	if err := http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatal(err)
	}
}
