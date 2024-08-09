package store

import (
	"time"
)

type User struct {
	ID       int64
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type House struct {
	HouseNumber     int64     `json:"house_number"`
	Address         string    `json:"address"`
	YearBuilt       int       `json:"year_built"`
	Developer       string    `json:"developer,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	LastFlatAddedAt time.Time `json:"last_flat_added_at,omitempty"`
}

type Flat struct {
	ID          int64
	HouseNumber int64  `json:"house_number"`
	FlatNumber  int64  `json:"flat_number"`
	Price       int    `json:"price"`
	Rooms       int    `json:"rooms"`
	Status      string `json:"status"`
	Moderator   string
}

type Database interface {
	CreateTable() error

	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)

	CreateHouse(house *House) error
	GetHouseByID(id int64) (*House, error)
	UpdateHouse(house *House) error
	UpdateHouseFlatTime(time time.Time) error

	CreateFlat(flat *Flat) error
	GetFlatsByHouseID(houseID int64, userType string) ([]Flat, error)
	UpdateFlat(flat *Flat, token string) error
	GetFlatStatus(houseID int64, flatNumber int64) (Flat, error)
	GetFlat(houseNumber, flatNumber int64) (*Flat, error)
}
