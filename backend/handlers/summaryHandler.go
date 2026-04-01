package handlers

import (
	"Backend/services"
	"net/http"
)

type SummaryHandler struct {
	Service *services.SummaryService
}

func (s *SummaryHandler) HandleAnalytics(w http.ResponseWriter, r *http.Request) {

}

func (s *SummaryHandler) HandleInsights(w http.ResponseWriter, r *http.Request) {

}
