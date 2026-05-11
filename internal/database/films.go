package database

import (
	"REST_api_appl/internal/models"
	"fmt"
	"gorm.io/gorm"
)

type FilmsStore struct {
	db *gorm.DB
}

func NewFilmStore(db *gorm.DB) *FilmsStore {
	return &FilmsStore{db: db}
}

func (s *FilmsStore) GetAll() ([]models.Film, error) {

	var films []models.Film

	result := s.db.Order("created_at DESC").Find(&films)

	if result.Error != nil {
		return nil, result.Error
	}

	return films, nil
}

func (s *FilmsStore) GetByIdAll(id int) (*models.Film, error) {
	var films models.Film

	result := s.db.Where("id = ?", id).Find(&films)

	if result.Error != nil {
		return nil, result.Error
	}

	return &films, nil
}

func (s *FilmsStore) Create(input models.CreateFilmInput) (*models.Film, error) {

	film := &models.Film{
		NameFilm:  input.NameFilm,
		Duration:  input.Duration,
		Completed: input.Completed,
	}

	err := s.db.Create(film)

	if err != nil {
		return nil, err.Error
	}

	return film, nil
}

func (s *FilmsStore) Update(id int, input models.UpdateFilmInput) (*models.Film, error) {
	film, err := s.GetByIdAll(id)

	if err := s.db.Model(&film).Updates(input).Error; err != nil {
		return nil, err
	}

	return film, err
}

func (s *FilmsStore) Delete(id int) error {
	result := s.db.Delete(&models.Film{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("Film with %d id not found", id)
	}

	return nil
}
