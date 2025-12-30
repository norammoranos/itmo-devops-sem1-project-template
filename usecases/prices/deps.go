package prices

import "project_sem/internal/infrastructure/database"

type Storage interface {
	database.Connection
}
