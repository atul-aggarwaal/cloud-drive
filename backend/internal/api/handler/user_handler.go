package handler

import (
	"encoding/json"
	"net/http"

	"github.com/atul-aggarwaal/cloud-drive/internal/usecase"
)

type UserHandler struct {
	userService *usecase.UserService
}

func NewUserHandler(service *usecase.UserService) *UserHandler {
	return &UserHandler{
		userService: service,
	}
}

// Struct representing Input Request
type RegisterUserRequest struct {
	UserName string `json:"username"`
	email    string `json:"email"`
	password string `json:"password"`
}

// Struct representing handler's response
type RegisterUserResponse struct {
	ID       string `json:"id"`
	UserName string `json:"username"`
	Email    string `json:"email"`
}

// Handles new user Registration. Validates user Inputs and create new Uer accordingly through UserService
func (this *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	//1. Make sure request comes from HTTP POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//2. Decode and parse input request
	var req RegisterUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Malformed Request", http.StatusBadRequest)
		return
	}

	// 3. Perform input validations
	if req.UserName == "" || req.email == "" || req.password == "" {
		http.Error(w, "Missing mandatory properties : User Name, Password and email", http.StatusBadRequest)
		return
	}
	//TODO email, username and Password length validations, password policy enforcement etc.

	//4. Send user details to UserService for user cration
	user, err := this.userService.RegisterUser(r.Context(), req.UserName, req.email, req.password)

	//5. Handle UserService errors, if any
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//6. Prepare and send response
	response := RegisterUserResponse{
		ID:       user.ID,
		UserName: user.UserName,
		Email:    user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
