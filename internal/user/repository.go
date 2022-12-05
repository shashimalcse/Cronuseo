package user

import (
	"context"

	"github.com/shashimalcse/cronuseo/internal/entity"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Get(ctx context.Context, org_id string, id string) (entity.User, error)
	Query(ctx context.Context, org_id string) ([]entity.User, error)
	Create(ctx context.Context, org_id string, user entity.User) error
	Update(ctx context.Context, org_id string, user entity.User) error
	Delete(ctx context.Context, org_id string, id string) error
	ExistByID(ctx context.Context, id string) (bool, error)
	ExistByKey(ctx context.Context, username string) (bool, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return repository{db: db}
}

func (r repository) Get(ctx context.Context, org_id string, id string) (entity.User, error) {
	user := entity.User{}
	err := r.db.Get(&user, "SELECT * FROM org_user WHERE org_id = $1 AND user_id = $2", org_id, id)
	return user, err
}

func (r repository) Create(ctx context.Context, org_id string, user entity.User) error {

	stmt, err := r.db.Prepare("INSERT INTO org_user(username,firstname,lastname,org_id,user_id) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.Username, user.FirstName, user.LastName, org_id, user.ID)
	if err != nil {
		return err
	}
	return nil

}

func (r repository) Update(ctx context.Context, org_id string, user entity.User) error {
	stmt, err := r.db.Prepare("UPDATE org_user SET firstname = $1, lastname = $2, WHERE org_id = $3 AND user_id = $4")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(user.FirstName, user.LastName, org_id, user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r repository) Delete(ctx context.Context, org_id string, id string) error {
	stmt, err := r.db.Prepare("DELETE FROM org_user WHERE org_id = $3 AND user_id = $1")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(org_id, id)
	if err != nil {
		return err
	}
	return nil
}

func (r repository) Query(ctx context.Context, org_id string) ([]entity.User, error) {
	users := []entity.User{}
	err := r.db.Select(&users, "SELECT * FROM org_user WHERE org_id = $1", org_id)
	return users, err
}

func (r repository) ExistByID(ctx context.Context, id string) (bool, error) {
	exists := false
	err := r.db.QueryRow("SELECT exists (SELECT user_id FROM org_user WHERE user_id = $1)", id).Scan(&exists)
	return exists, err
}

func (r repository) ExistByKey(ctx context.Context, username string) (bool, error) {
	exists := false
	err := r.db.QueryRow("SELECT exists (SELECT user_id FROM org_user WHERE username = $1)", username).Scan(&exists)
	return exists, err
}