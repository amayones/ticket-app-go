package routes

import (
	"time"

	"github.com/go-chi/chi/v5"

	"golang-backend/handlers"
	"golang-backend/middleware"
)

func SetupRoutes(userHandler *handlers.UserHandler) *chi.Mux {
	r := chi.NewRouter()
	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	r.Route("/api", func(r chi.Router) {
		r.With(loginLimiter.Middleware).Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.RefreshToken)
		r.Post("/logout", userHandler.Logout)
		r.Route("/users", func(r chi.Router) {
			r.With(middleware.AuthMiddleware).Get("/", userHandler.GetUsers)
			r.Post("/", userHandler.CreateUser)
			r.With(middleware.AuthMiddleware).Get("/{id}", userHandler.GetUserByID)
			r.With(middleware.AuthMiddleware).Put("/{id}", userHandler.UpdateUser)
			r.With(middleware.AuthMiddleware).Delete("/{id}", userHandler.DeleteUser)
		})
	})
	return r
}
