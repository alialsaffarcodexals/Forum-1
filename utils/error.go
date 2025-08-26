package utils

import (
	"html/template"
	"log"
	"net/http"
)

func RenderError(w http.ResponseWriter, message string, statusCode int) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		log.Println("Template parse error in RenderError:", err)
		return
	}
	w.WriteHeader(statusCode)
	err = tmpl.Execute(w, map[string]interface{}{
		"StatusCode": statusCode,
		"Message":    message,
	})
	if err != nil {
		log.Println("Template execute error in RenderError:", err)
	}
}
