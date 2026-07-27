package model

import "time"

type Url struct {
	Id        int       `json:"id" db:"id"`
	Url       string    `json:"url" db:"url"`
	Alias     string    `json:"alias" db:"alias"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
