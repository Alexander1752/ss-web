package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/mongo"

	"mqtt-streaming-server/repository"
)

// UserController handles HTTP requests for user operations.
// Business logic is delegated to UserService.
type UserController struct {
	Service *UserService
}

func InitUserRoutes(db *mongo.Database, mux *http.ServeMux) {
	userController := &UserController{
		Service: NewUserService(repository.NewUserRepository(db)),
	}

	mux.HandleFunc("/register", userController.Register)
	mux.HandleFunc("/login", userController.Login)
	mux.Handle("/profile", withAuth(http.HandlerFunc(userController.GetProfile)))
}

func (ctlr UserController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := ctlr.Service.Register(r.Context(), req.Email, req.Password); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			http.Error(w, "User already exists", http.StatusConflict)
		} else if errors.Is(err, errCheckExistingUser) {
			http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
		} else {
			http.Error(w, "Failed to save user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User registered successfully")
}

func (ctlr UserController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tokenStr, user, err := ctlr.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Invalid email or password: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":   tokenStr,
		"message": "Login successful",
		"email":   user.Email,
		"role":    user.Role,
	})
}

func (ctlr UserController) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email, ok := r.Context().Value("email").(string)
	if !ok {
		http.Error(w, "Email not found in context", http.StatusUnauthorized)
		return
	}

	user, err := ctlr.Service.GetProfile(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
