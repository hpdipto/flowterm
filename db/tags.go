package db

import (
	"database/sql"
	"errors"
	"regexp"
	"time"
)

var tagNameRegex = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,28}[a-z0-9])?$`)

func validateTagName(name string) error {
	if !tagNameRegex.MatchString(name) {
		return errors.New("invalid tag name")
	}

	return nil
}

func (d *DB) GetTagByName(name string) (*Tag, error) {
	tag := &Tag{}

	err := d.conn.QueryRow(`
		SELECT id, name, created_at, updated_at
		FROM tags
		WHERE name = ?
	`, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (d *DB) CreateTag(tag *Tag) error {
	now := time.Now()

	result, err := d.conn.Exec(`
		INSERT INTO tags (
			name,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?)
	`,
		tag.Name,
		Timestamp(now),
		Timestamp(now),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	tag.ID = id
	tag.CreatedAt = now
	tag.UpdatedAt = now

	return nil
}

func (d *DB) GetOrCreateTag(name string) (*Tag, error) {
	if err := validateTagName(name); err != nil {
		return nil, err
	}

	tag, err := d.GetTagByName(name)
	if err == nil {
		return tag, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	tag = &Tag{Name: name}

	if err := d.CreateTag(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (d *DB) AddTagToSession(sessionID, tagID int64) error {
	_, err := d.conn.Exec(`
		INSERT OR IGNORE INTO session_tags (
			session_id,
			tag_id,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?)
	`,
		sessionID,
		tagID,
		Timestamp(time.Now()),
		Timestamp(time.Now()),
	)

	return err
}
