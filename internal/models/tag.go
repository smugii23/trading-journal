package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Tag struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Category string `json:"category"`
	Color    string `json:"color"`
}

func CreateTag(db *sql.DB, tag *Tag) error {
	stmt, err := db.Prepare("INSERT INTO tags (user_id, category, color) VALUES ($1, $2, $3) RETURNING id")
	if err != nil {
		log.Printf("error creating tag: %v", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}
	_, err = stmt.Exec(tag.UserID, tag.Category, tag.Color)
	if err != nil {
		log.Printf("error creating tag: %v", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func GetTagsByUserID(db *sql.DB, userID int) ([]Tag, error) {
	rows, err := db.Query("SELECT id, user_id, category, color FROM tags WHERE user_id = $1", userID)
	if err != nil {
		log.Printf("error retrieving tags: %v", err)
		return nil, fmt.Errorf("error retrieving tag: %w", err)
	}
	defer rows.Close()
	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Category, &t.Color); err != nil {
			return nil, fmt.Errorf("error scanning tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}
	return tags, nil
}
