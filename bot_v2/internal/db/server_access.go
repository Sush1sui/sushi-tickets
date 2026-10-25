package db

import "context"

// ServerAuthorizedRole is one (server, role) whitelist entry.
type ServerAuthorizedRole struct {
	ServerConfigID int64
	RoleID         string
}

const getExistingServerConfigIDs = `-- name: GetExistingServerConfigIDs :many
SELECT id FROM server_config WHERE id = ANY($1::bigint[])
`

func (q *Queries) GetExistingServerConfigIDs(ctx context.Context, serverIDs []int64) ([]int64, error) {
	rows, err := q.db.Query(ctx, getExistingServerConfigIDs, serverIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

const getAuthorizedRolesForServers = `-- name: GetAuthorizedRolesForServers :many
SELECT server_config_id, role_id FROM authorized_roles
WHERE server_config_id = ANY($1::bigint[])
`

func (q *Queries) GetAuthorizedRolesForServers(ctx context.Context, serverIDs []int64) ([]ServerAuthorizedRole, error) {
	rows, err := q.db.Query(ctx, getAuthorizedRolesForServers, serverIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ServerAuthorizedRole
	for rows.Next() {
		var i ServerAuthorizedRole
		if err := rows.Scan(&i.ServerConfigID, &i.RoleID); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
