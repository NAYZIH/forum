package functions

import (
	"database/sql"
	"forum/backend/database"
	"forum/backend/models"
)

func CreatePost(userID int, title, content string, categories []string, imagePath string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	res, err := tx.Exec("INSERT INTO posts (user_id, title, content, image_path) VALUES (?, ?, ?, ?)", userID, title, content, imagePath)
	if err != nil {
		tx.Rollback()
		return err
	}
	postID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, cat := range categories {
		if cat == "" {
			continue
		}
		var catID int
		err = tx.QueryRow("SELECT id FROM categories WHERE name = ?", cat).Scan(&catID)
		if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", cat)
			if err != nil {
				tx.Rollback()
				return err
			}
			catID64, err := res.LastInsertId()
			if err != nil {
				tx.Rollback()
				return err
			}
			catID = int(catID64)
		} else if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func UpdatePost(postID int, title, content string, categories []string, imagePath string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE posts SET title = ?, content = ?, image_path = ? WHERE id = ?", title, content, imagePath, postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, cat := range categories {
		if cat == "" {
			continue
		}
		var catID int
		err = tx.QueryRow("SELECT id FROM categories WHERE name = ?", cat).Scan(&catID)
		if err == sql.ErrNoRows {
			res, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", cat)
			if err != nil {
				tx.Rollback()
				return err
			}
			catID64, err := res.LastInsertId()
			if err != nil {
				tx.Rollback()
				return err
			}
			catID = int(catID64)
		} else if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func GetPostByID(id int) (*models.Post, error) {
	row := database.DB.QueryRow(`
        SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        WHERE p.id = ?
    `, id)
	post := &models.Post{}
	err := row.Scan(&post.ID, &post.UserID, &post.Username, &post.Title, &post.Content, &post.ImagePath, &post.CreatedAt)
	if err != nil {
		return nil, err
	}
	cats, err := GetCategoriesByPostID(post.ID)
	if err != nil {
		return nil, err
	}
	post.Categories = cats
	err = database.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND value = 1", post.ID).Scan(&post.Likes)
	if err != nil {
		return nil, err
	}
	err = database.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND value = -1", post.ID).Scan(&post.Dislikes)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func GetPosts(filter, category string, userID int) ([]models.Post, error) {
	var query string
	var args []interface{}
	switch filter {
	case "category":
		query = `
            SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
            FROM posts p
            JOIN users u ON p.user_id = u.id
            JOIN post_categories pc ON p.id = pc.post_id
            JOIN categories c ON pc.category_id = c.id
            WHERE c.name = ?
            ORDER BY p.created_at DESC
        `
		args = []interface{}{category}
	case "created":
		query = `
            SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
            FROM posts p
            JOIN users u ON p.user_id = u.id
            WHERE p.user_id = ?
            ORDER BY p.created_at DESC
        `
		args = []interface{}{userID}
	case "liked":
		query = `
            SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
            FROM posts p
            JOIN users u ON p.user_id = u.id
            JOIN post_likes pl ON p.id = pl.post_id
            WHERE pl.user_id = ? AND pl.value = 1
            ORDER BY p.created_at DESC
        `
		args = []interface{}{userID}
	default:
		query = `
            SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
            FROM posts p
            JOIN users u ON p.user_id = u.id
            ORDER BY p.created_at DESC
        `
	}
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		cats, err := GetCategoriesByPostID(p.ID)
		if err != nil {
			return nil, err
		}
		p.Categories = cats
		err = database.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND value = 1", p.ID).Scan(&p.Likes)
		if err != nil {
			return nil, err
		}
		err = database.DB.QueryRow("SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND value = -1", p.ID).Scan(&p.Dislikes)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func GetLikedPostsByUserID(userID int) ([]models.Post, error) {
	rows, err := database.DB.Query(`
        SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN post_likes pl ON p.id = pl.post_id
        WHERE pl.user_id = ? AND pl.value = 1
        ORDER BY pl.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		cats, err := GetCategoriesByPostID(p.ID)
		if err != nil {
			return nil, err
		}
		p.Categories = cats
		posts = append(posts, p)
	}
	return posts, nil
}

func GetDislikedPostsByUserID(userID int) ([]models.Post, error) {
	rows, err := database.DB.Query(`
        SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN post_likes pl ON p.id = pl.post_id
        WHERE pl.user_id = ? AND pl.value = -1
        ORDER BY pl.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var p models.Post
		err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		cats, err := GetCategoriesByPostID(p.ID)
		if err != nil {
			return nil, err
		}
		p.Categories = cats
		posts = append(posts, p)
	}
	return posts, nil
}

func DeletePost(postID int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM comment_likes WHERE comment_id IN (SELECT id FROM comments WHERE post_id = ?)", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM comments WHERE post_id = ?", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM post_likes WHERE post_id = ?", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM post_categories WHERE post_id = ?", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
