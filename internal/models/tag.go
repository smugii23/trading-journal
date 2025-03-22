package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Tag struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
}

func CreateTag(db *sql.DB, tag *Tag) error {
	stmt, err := db.Prepare("INSERT INTO tags (user_id, category, color) VALUES ($1, $2, $3) RETURNING id")
	if err != nil {
		log.Printf("error creating tag: %v", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}
	err = stmt.QueryRow(tag.UserID, tag.Category, tag.Color).Scan(&tag.ID)
	if err != nil {
		log.Printf("error creating tag: %v", err)
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

func GetTagsByUserID(db *sql.DB, userID int) ([]Tag, error) {
	rows, err := db.Query("SELECT id, user_id, name, category, color FROM tags WHERE user_id = $1", userID)
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

func AddTagToTrade(db *sql.DB, tradeID, tagID int) error {
	_, err := db.Exec("INSERT INTO trade_tags (trade_id, tag_id) VALUES ($1, $2)", tradeID, tagID)
	if err != nil {
		return fmt.Errorf("error adding tag to trade: %w", err)
	}
	return nil
}

func RemoveTagFromTrade(db *sql.DB, tradeID, tagID int) error {
	_, err := db.Exec("DELETE FROM trade_tags WHERE trade_id = $1 AND tag_id = $2", tradeID, tagID)
	if err != nil {
		return fmt.Errorf("error deleting tag from trade: %w", err)
	}
	return nil
}

func GetTagsFromTrade(db *sql.DB, tradeID int) ([]Tag, error) {
	rows, err := db.Query("SELECT t.id, t.user_id, t.name, t.category, t.color FROM tags t JOIN trade_tags tt ON t.id = tt.tag_id WHERE tt.trade_id = $1", tradeID)
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
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
