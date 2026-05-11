package handlers

import (
	"REST_api_appl/internal/cache"
	"REST_api_appl/internal/database"
	"REST_api_appl/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HandlersTask struct {
	store *database.FilmsStore
	cache *cache.FilmsCache
}

func NewHandlers(store *database.FilmsStore, filmsCache *cache.FilmsCache) *HandlersTask {
	return &HandlersTask{store: store, cache: filmsCache}
}

func ResponseWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func ResponseWithError(w http.ResponseWriter, statusCode int, message string) {
	ResponseWithJSON(w, statusCode, map[string]string{"error": message})
}

func parseTaskIDFromPath(path string) (int, error) {
	pathPats := strings.Split(strings.TrimPrefix(path, "/tasks/"), "/")
	if len(pathPats) == 0 || pathPats[0] == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(pathPats[0])
}

func (h *HandlersTask) GetAllFilms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Пробуем взять из кэша
	films, err := h.cache.GetList(ctx)
	if err == nil {
		ResponseWithJSON(w, http.StatusOK, films)
		log.Println("Данные успешно взяты из кэша")
		return
	}
	// redis.Nil = ключа нет, идём в БД; иначе — ошибка кэша, тоже идём в БД

	films, err = h.store.GetAll()
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, "Ошибка в получении задач")
		return
	}

	_ = h.cache.SetList(ctx, films)

	ResponseWithJSON(w, http.StatusOK, films)
}

func (h *HandlersTask) GetFilms(w http.ResponseWriter, r *http.Request) {

	id, err := parseTaskIDFromPath(r.URL.Path)

	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, "Ошибка получения задачи не тот ID")
		return
	}

	film, _ := h.store.GetByIdAll(id)

	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ResponseWithJSON(w, http.StatusOK, &film)
}

func (h *HandlersTask) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input models.CreateFilmInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		ResponseWithError(w, http.StatusBadRequest, "Неккоректно отправленные данные")
		return
	}

	if strings.TrimSpace(input.NameFilm) == "" {
		ResponseWithError(w, http.StatusBadRequest, "Заголовок тоже важен")
		return
	}

	task, err := h.store.Create(input)
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = h.cache.Invalidate(r.Context()) // инвалидируем кэш при изменении данных
	ResponseWithJSON(w, http.StatusOK, task)
}

func (h *HandlersTask) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskIDFromPath(r.URL.Path)
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, "Ошибка получения задачи не тот ID")
		return
	}

	var input models.UpdateFilmInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		ResponseWithError(w, http.StatusBadRequest, "Неверно отправленные данные")
		return
	}

	if input.NameFilm != nil && strings.TrimSpace(*input.NameFilm) == "" {
		ResponseWithError(w, http.StatusBadRequest, "Заголовок тоже важен")
		return
	}

	task, err := h.store.Update(id, input)
	if err != nil {
		ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_ = h.cache.Invalidate(r.Context())
	ResponseWithJSON(w, http.StatusOK, task)
}

func (h *HandlersTask) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskIDFromPath(r.URL.Path)
	if err != nil {
		ResponseWithError(w, http.StatusBadRequest, "Не тот ID")
		return
	}

	err = h.store.Delete(id)

	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			ResponseWithError(w, http.StatusBadRequest, err.Error())
		} else {
			ResponseWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	_ = h.cache.Invalidate(r.Context())
	ResponseWithJSON(w, http.StatusOK, map[string]string{"message": "success"})
}
