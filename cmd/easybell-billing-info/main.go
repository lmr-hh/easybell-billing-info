package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lmr-hh/functions/CurrentUsage"
	"github.com/lmr-hh/functions/MonthlyUsage"
)

func main() {
	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}
	timeZone, err := time.LoadLocation(os.Getenv("WEBSITE_TIME_ZONE"))
	if err != nil {
		log.Fatalln(err)
	}
	username, found := os.LookupEnv("EASYBELL_USERNAME")
	if !found {
		log.Fatalln("EASYBELL_USERNAME is required")
	}
	password, found := os.LookupEnv("EASYBELL_PASSWORD")
	if !found {
		log.Fatalln("EASYBELL_PASSWORD is required")
	}
	url := os.Getenv("EASYBELL_TEAMS_WEBHOOK")
	if url == "" {
		log.Fatalln("EASYBELL_TEAMS_WEBHOOK is required")
	}

	currentUsageHandler := CurrentUsage.NewHandler(username, password, timeZone, url)
	monthlyUsageHandler := MonthlyUsage.NewHandler(username, password, timeZone, url)

	minutes, err := strconv.Atoi(os.Getenv("EASYBELL_NATIONAL_MINUTES"))
	if err != nil {
		log.Fatalln("EASYBELL_NATIONAL_MINUTES must be an integer.")
	}
	currentUsageHandler.NationalMinutes = minutes
	monthlyUsageHandler.NationalMinutes = minutes
	minutes, err = strconv.Atoi(os.Getenv("EASYBELL_MOBILE_MINUTES"))
	if err != nil {
		log.Fatalln("EASYBELL_MOBILE_MINUTES must be an integer")
	}
	currentUsageHandler.MobileMinutes = minutes
	monthlyUsageHandler.MobileMinutes = minutes

	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/CurrentUsage", currentUsageHandler)
	http.Handle("/MonthlyUsage", monthlyUsageHandler)
	log.Printf("Listening on http://127.0.0.1%s/", listenAddr)
	log.Fatalln(http.ListenAndServe(listenAddr, nil))

}
