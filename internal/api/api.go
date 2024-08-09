package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"avtest/internal/store"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

const (
	Client    = "client"
	Moderator = "moderator"
)

var (
	lock      sync.Mutex
	jwtKey    = []byte("secret-key")
	statuses  = []string{"on moderation", "approved", "declined"}
	userTypes = []string{Client, Moderator}

	errFailedToUpdateFlat   = errors.New("another moderator has already been assigned to this flat")
	errInvalidSigningMethod = errors.New("unexpected signing method")
	errInvalidToken         = errors.New("invalid token")
	errInvalidUserType      = errors.New("invalid user type")
	errFailedToGenerateJWT  = errors.New("failed to generate jwt")
	errUserExists           = errors.New("user already exists")
	errNonExistentUser      = errors.New("user doesn't exists")
	errFailedToCheckToken   = errors.New("failed to check token")
	errUnauthorized         = errors.New("not authorized")
	errFailedToSubscribe    = errors.New("subscribe for clients only")
	errWhongStatus          = errors.New("wrong status for the flat")
)

type Claims struct {
	Role string
	jwt.StandardClaims
}

type API struct {
	logger *zap.Logger
	r      *mux.Router
	db     store.Database
}

func NewAPI(logger *zap.Logger, r *mux.Router, db store.Database) *API {
	return &API{
		logger: logger,
		r:      r,
		db:     db,
	}
}

func (a *API) Run(apiPort string) {
	a.r.HandleFunc("/dummyLogin", a.dummyLoginHandler).Methods("POST")
	a.r.HandleFunc("/register", a.registerHandler).Methods("POST")
	a.r.HandleFunc("/login", a.loginHandler).Methods("POST")
	a.r.HandleFunc("/house/create", a.createHouseHandler).Methods("POST")
	a.r.HandleFunc("/flat/create", a.createFlatHandler).Methods("POST")
	a.r.HandleFunc("/flat/update", a.updateFlatHandler).Methods("POST")
	a.r.HandleFunc("/house/{id:[a-zA-Z0-9]+}", a.getFlatsByHouseHandler).Methods("GET")
	a.r.HandleFunc("/house/{id:[a-zA-Z0-9]+}/subscribe", a.subscribeHandler).Methods("POST")

	http.Handle("/", handlers.CORS(handlers.AllowedOrigins([]string{"*"}))(a.r))
	http.ListenAndServe(apiPort, nil)
}

func (a *API) dummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req *store.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userType := req.Type
	if !slices.Contains(userTypes, userType) {
		http.Error(w, errInvalidUserType.Error(), http.StatusBadRequest)
		return
	}

	token, err := generateToken(userType)
	if err != nil {
		http.Error(w, errFailedToGenerateJWT.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *API) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req *store.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	u, err := a.db.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u != nil {
		http.Error(w, errUserExists.Error(), http.StatusNotFound)
		return
	}

	err = a.db.CreateUser(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "user created"})
}

func (a *API) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req *store.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	u, err := a.db.GetUserByEmail(req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u == nil {
		http.Error(w, errNonExistentUser.Error(), http.StatusNotFound)
		return
	}

	token, err := generateToken(u.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *API) createHouseHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	err := checkModerator(bearerToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req *store.House
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	req.CreatedAt = time.Now()

	err = a.db.CreateHouse(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

func checkModerator(bearerToken string) error {
	tokenString := getCorrectToken(bearerToken)
	token, err := checkToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to check token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Role != "moderator" {
		return fmt.Errorf("not authorized")
	}
	return nil
}

func (a *API) createFlatHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	err := checkModerator(bearerToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req *store.Flat
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	req.Status = "created"

	err = a.db.CreateFlat(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = a.db.UpdateHouseFlatTime(time.Now())
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

func (a *API) updateFlatHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	tokenString := getCorrectToken(bearerToken)
	err := checkModerator(bearerToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req *store.Flat
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Lock()
	defer lock.Unlock()

	f, err := a.db.GetFlatStatus(req.HouseNumber, req.FlatNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	curStatus := f.Status
	if curStatus == "on moderation" && f.Moderator != tokenString {
		http.Error(w, errFailedToUpdateFlat.Error(), http.StatusBadRequest)
		return
	}

	if !slices.Contains(statuses, req.Status) {
		http.Error(w, errWhongStatus.Error(), http.StatusBadRequest)
		return
	}

	err = a.db.UpdateFlat(req, tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	flat, err := a.db.GetFlat(req.HouseNumber, req.FlatNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(flat)
}

func (a *API) getFlatsByHouseHandler(w http.ResponseWriter, r *http.Request) {
	houseID := mux.Vars(r)["id"]
	bearerToken := r.Header.Get("Authorization")
	tokenString := getCorrectToken(bearerToken)
	userType, err := getUserType(tokenString)
	if err != nil {
		http.Error(w, "failed to get user type", http.StatusBadRequest)
	}

	lock.Lock()
	defer lock.Unlock()

	id, err := strconv.ParseInt(houseID, 10, 2)
	if err != nil {
		log.Fatalf("failed to parse int: %s", err)
	}

	h, err := a.db.GetHouseByID(id)
	if err != nil {
		http.Error(w, "Flat not found", http.StatusNotFound)
		return
	}
	if h == nil {
		http.Error(w, "House not found", http.StatusNotFound)
		return
	}

	flats, err := a.db.GetFlatsByHouseID(id, userType)
	if err != nil {
		http.Error(w, "Flats not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(flats)
}

func (a *API) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	tokenString := getCorrectToken(bearerToken)
	userType, err := getUserType(tokenString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if userType != "client" {
		http.Error(w, errFailedToSubscribe.Error(), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscribed successfully"})
}

// generateToken generates a token based on the user type.
func generateToken(role string) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)

	tokenClaims := &Claims{
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString(jwtKey)
}

// checkToken checks whether the token meets the requirements.
func checkToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("%w: %v", errInvalidSigningMethod, err)
	}
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("%w: %v", errInvalidToken, err)
	}
	return token, nil
}

// getCorrectToken converts the token to the correct form.
func getCorrectToken(token string) string {
	splitToken := strings.Split(token, " ")
	var correctToken string
	if len(splitToken) == 2 {
		correctToken = splitToken[1]
	}
	return correctToken
}

// getUserType returns the user type encrypted in the token.
func getUserType(tokenString string) (string, error) {
	token, err := checkToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("%w: %v", errFailedToCheckToken, err)
	}

	tokenClaims, ok := token.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("%w: %v", errUnauthorized, err)
	}

	return tokenClaims.Role, nil
}
