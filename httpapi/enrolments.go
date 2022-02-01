package httpapi

import "net/http"

func getEnrolment(api enrolmentsAPI) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		// TODO: implement this.
	})
}

func listEnrolments(api enrolmentsAPI) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		// TODO: implement this.
	})
}

func enrol(api enrolmentsAPI) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		// TODO: implement this.
	})
}

func ingest(api enrolmentsAPI) http.HandlerFunc {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		// TODO: implement this.
	})
}
