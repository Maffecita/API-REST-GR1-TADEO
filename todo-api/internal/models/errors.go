package models

import "errors"

// Errores del dominio (modelos)
var (
	ErrEmptyTitle      = errors.New("el título no puede estar vacío")
	ErrTitleTooLong    = errors.New("el título no puede superar 150 caracteres")
	ErrInvalidPriority = errors.New("prioridad inválida: usa 'baja', 'media' o 'alta'")
	ErrMissingFields   = errors.New("username, email y password son obligatorios")
	ErrPasswordTooShort = errors.New("la contraseña debe tener al menos 6 caracteres")
)

// Errores de repositorio
var (
	ErrNotFound      = errors.New("recurso no encontrado")
	ErrDuplicate     = errors.New("el recurso ya existe")
	ErrUnauthorized  = errors.New("no autorizado")
)
