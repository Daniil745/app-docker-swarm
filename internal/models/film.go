package models

import "time"

type Film struct {
	ID        uint      `gorm:"primarykey;column:id" json:"id"`
	NameFilm  string    `gorm:"column:name" json:"nameFilm"`
	Duration  int       `gorm:"column:duration" json:"duration"`
	Completed bool      `gorm:"column:completed" json:"completed"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type CreateFilmInput struct {
	NameFilm  string `json:"nameFilm"`
	Duration  int    `json:"duration"`
	Completed bool   `json:"completed"`
}

type UpdateFilmInput struct {
	NameFilm  *string `json:"nameFilm"`
	Duration  *int    `json:"duration"`
	Completed *bool   `json:"completed"`
}
