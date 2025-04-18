package functions

import "forum/backend/database"

func GetAllCategories() ([]string, error) {
	rows, err := database.DB.Query("SELECT name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var cat string
		err := rows.Scan(&cat)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func GetCategoriesByPostID(postID int) ([]string, error) {
	rows, err := database.DB.Query("SELECT c.name FROM categories c JOIN post_categories pc ON c.id = pc.category_id WHERE pc.post_id = ?", postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var cat string
		err := rows.Scan(&cat)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

func DeleteOrphanCategories() error {
	_, err := database.DB.Exec(`
		DELETE FROM categories
		WHERE id NOT IN (SELECT DISTINCT category_id FROM post_categories)
	`)
	return err
}
