package http_request

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yoonjin67/linux_virt_unit"
	"github.com/yoonjin67/linux_virt_unit/incus_unit"
)


func InitHttpRequest() {
	linux_virt_unit.LinuxVirtualizationAPIRouter = mux.NewRouter()
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/register", incus_unit.Register).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/create", incus_unit.CreateContainer).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/request", incus_unit.GetContainers).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/delete", incus_unit.DeleteByTag).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/stop", incus_unit.StopByTag).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/start", incus_unit.StartByTag).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/pause", incus_unit.PauseByTag).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/resume", incus_unit.ResumeByTag).Methods("POST")
	linux_virt_unit.LinuxVirtualizationAPIRouter.HandleFunc("/restart", incus_unit.RestartByTag).Methods("POST")

	srv := &http.Server{
		Handler:      linux_virt_unit.LinuxVirtualizationAPIRouter,
		Addr:         ":32000",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on port 32000")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe error: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully.")
	}
}
