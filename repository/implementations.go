package repository

import (
	"context"
	"database/sql"
	"fmt"
)

func (r *Repository) CreateEstate(ctx context.Context, input Estate) (result Estate, err error) {
	query := `
        INSERT INTO estates (id, width, length)
        VALUES ($1, $2, $3)
        RETURNING id;
    `
	err = r.Db.QueryRowContext(ctx, query, input.Id, input.Width, input.Length).Scan(&result.Id)
	if err != nil {
		return result, fmt.Errorf("failed to create estate: %w", err)
	}

	result.Id = input.Id
	result.Width = input.Width
	result.Length = input.Length
	return result, nil
}

func (r *Repository) CreateEstateTree(ctx context.Context, input EstateTree) (result EstateTree, err error) {
	query := `
        INSERT INTO trees (id, estate_id, x, y, height)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id;
    `
	err = r.Db.QueryRowContext(ctx, query, input.Id, input.EstateId, input.X, input.Y, input.Height).Scan(&result.Id)
	if err != nil {
		return result, fmt.Errorf("failed to create tree: %w", err)
	}

	result.EstateId = input.EstateId
	result.X = input.X
	result.Y = input.Y
	result.Height = input.Height
	return result, nil
}

func (r *Repository) GetStatsByEstateId(ctx context.Context, id string) (result StatsEstate, err error) {
	query := `
        SELECT 
            COALESCE(COUNT(*), 0) AS count, 
            COALESCE(MAX(height), 0) AS max_height, 
            COALESCE(MIN(height), 0) AS min_height, 
            COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY height), 0) AS median_height
        FROM trees
        WHERE estate_id = $1;
    `
	err = r.Db.QueryRowContext(ctx, query, id).Scan(
		&result.Count,
		&result.Max,
		&result.Min,
		&result.Median,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, fmt.Errorf("no statistics found for estate_id %s: %w", id, err)
		}
		return result, fmt.Errorf("failed to retrieve stats: %w", err)
	}
	return result, nil
}

func (r *Repository) GetEstateById(ctx context.Context, id string) (result Estate, err error) {
	query := `
        SELECT id, width, length 
        FROM estates 
        WHERE id = $1;
    `
	err = r.Db.QueryRowContext(ctx, query, id).Scan(&result.Id, &result.Width, &result.Length)
	if err != nil {
		if err == sql.ErrNoRows {
			return result, fmt.Errorf("estate with id %s not found: %w", id, err)
		}
		return result, fmt.Errorf("failed to retrieve estate: %w", err)
	}
	return result, nil
}

func (r *Repository) GetTreesByEstateId(ctx context.Context, id string) (result []EstateTree, err error) {
	query := `
        SELECT id, estate_id, x, y, height 
        FROM trees 
        WHERE estate_id = $1;
    `
	rows, err := r.Db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query trees: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close rows: %w", cerr)
		}
	}()

	for rows.Next() {
		var tree EstateTree
		if err = rows.Scan(&tree.Id, &tree.EstateId, &tree.X, &tree.Y, &tree.Height); err != nil {
			return nil, fmt.Errorf("failed to scan tree data: %w", err)
		}
		result = append(result, tree)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}
