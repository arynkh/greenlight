package data

import (
	"context"
	"database/sql"
	"slices"
	"time"
)

// Define a Permissions slice, which we will use to hold the permission codes (such as
// "movies:read" and "movies:write") for a single user.
type Permissions []string

// Add a helper method to check whether the Permissions slice contains a specific permission code
func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		select permissions.code
		from permissions
		inner join users_permissions on users_permissions.permission_id = permissions.id
		inner join users on users_permissions.user_id = users.id
		where users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions

	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}
