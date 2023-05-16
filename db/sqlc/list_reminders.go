package db

import (
	"context"
	"fmt"

	"github.com/OCD-Labs/KeyKeeper/internal/pagination"
)

type ListRemindersParamsX struct {
	WebsiteURL string
	Filters    pagination.Filters
}

// ListRemindersX do a fulltext search to list reminders, and paginates accordingly.
func (q *SQLStore) ListRemindersX(ctx context.Context, arg ListRemindersParamsX) ([]Reminder, pagination.Metadata, error) {
	stmt := fmt.Sprintf(`
		SELECT count(*) OVER() AS total_count, id, user_id, website_url, "interval", updated_at, extension
		FROM reminders
		WHERE (website_url LIKE '%%' || $1 || '%%' OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, arg.Filters.SortColumn(), arg.Filters.SortDirection(),
	)

	args := []interface{}{arg.WebsiteURL, arg.Filters.Limit(), arg.Filters.Offset()}

	rows, err := q.db.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, pagination.Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	items := []Reminder{}

	for rows.Next() {
		var i Reminder
		if err := rows.Scan(
			&totalRecords,
			&i.ID,
			&i.UserID,
			&i.WebsiteUrl,
			&i.Interval,
			&i.UpdatedAt,
			&i.Extension,
		); err != nil {
			return nil, pagination.Metadata{}, err
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, pagination.Metadata{}, err
	}

	metadata := pagination.CalcMetadata(totalRecords, arg.Filters.Page, arg.Filters.PageSize)

	return items, metadata, nil
}
