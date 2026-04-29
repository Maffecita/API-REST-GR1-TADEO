package repository

import (
	"database/sql"
	"time"
	"todo-api/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepository implementa el acceso a datos usando SQLite
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository crea e inicializa la base de datos
func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, err
	}
	return repo, nil
}

// migrate crea las tablas si no existen
func (r *SQLiteRepository) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			username   TEXT    NOT NULL UNIQUE,
			email      TEXT    NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id     INTEGER NOT NULL,
			title       TEXT    NOT NULL,
			description TEXT    DEFAULT '',
			priority    TEXT    DEFAULT 'media',
			completed   BOOLEAN DEFAULT FALSE,
			due_date    DATETIME,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
	}
	for _, q := range queries {
		if _, err := r.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

// ─── Métodos de Usuario ───────────────────────────────────────────────────────

func (r *SQLiteRepository) CreateUser(user *models.User, passwordHash string) error {
	query := `INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`
	result, err := r.db.Exec(query, user.Username, user.Email, passwordHash)
	if err != nil {
		return models.ErrDuplicate
	}
	id, _ := result.LastInsertId()
	user.ID = int(id)
	return nil
}

func (r *SQLiteRepository) GetUserByEmail(email string) (*models.User, string, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var user models.User
	var hash string
	err := row.Scan(&user.ID, &user.Username, &user.Email, &hash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, "", models.ErrNotFound
	}
	return &user, hash, err
}

func (r *SQLiteRepository) GetUserByID(id int) (*models.User, error) {
	query := `SELECT id, username, email, created_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	return &user, err
}

// ─── Métodos de Tareas ────────────────────────────────────────────────────────

func (r *SQLiteRepository) GetAllTasks(userID int) ([]models.Task, error) {
	query := `SELECT id, user_id, title, description, priority, completed, due_date, created_at, updated_at
	          FROM tasks WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var dueDate sql.NullTime
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description,
			&t.Priority, &t.Completed, &dueDate, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if dueDate.Valid {
			t.DueDate = &dueDate.Time
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *SQLiteRepository) GetTaskByID(id, userID int) (*models.Task, error) {
	query := `SELECT id, user_id, title, description, priority, completed, due_date, created_at, updated_at
	          FROM tasks WHERE id = ? AND user_id = ?`
	row := r.db.QueryRow(query, id, userID)

	var t models.Task
	var dueDate sql.NullTime
	err := row.Scan(&t.ID, &t.UserID, &t.Title, &t.Description,
		&t.Priority, &t.Completed, &dueDate, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, models.ErrNotFound
	}
	if dueDate.Valid {
		t.DueDate = &dueDate.Time
	}
	return &t, err
}

func (r *SQLiteRepository) CreateTask(task *models.Task) error {
	query := `INSERT INTO tasks (user_id, title, description, priority, due_date)
	          VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, task.UserID, task.Title, task.Description,
		task.Priority, task.DueDate)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	task.ID = int(id)
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return nil
}

func (r *SQLiteRepository) UpdateTask(task *models.Task) error {
	query := `UPDATE tasks SET title=?, description=?, priority=?, due_date=?, updated_at=CURRENT_TIMESTAMP
	          WHERE id=? AND user_id=?`
	result, err := r.db.Exec(query, task.Title, task.Description, task.Priority,
		task.DueDate, task.ID, task.UserID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *SQLiteRepository) DeleteTask(id, userID int) error {
	result, err := r.db.Exec(`DELETE FROM tasks WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *SQLiteRepository) MarkTaskComplete(id, userID int) (*models.Task, error) {
	_, err := r.db.Exec(
		`UPDATE tasks SET completed=TRUE, updated_at=CURRENT_TIMESTAMP WHERE id=? AND user_id=?`,
		id, userID,
	)
	if err != nil {
		return nil, err
	}
	return r.GetTaskByID(id, userID)
}
