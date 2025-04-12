package functions

import (
	"forum/backend/database"
	"forum/backend/models"
)

func CreateReport(moderatorID int, postID, commentID *int, reason string) error {
	_, err := database.DB.Exec(
		"INSERT INTO reports (moderator_id, post_id, comment_id, reason) VALUES (?, ?, ?, ?)",
		moderatorID, postID, commentID, reason,
	)
	return err
}

func GetPendingReports() ([]models.Report, error) {
	rows, err := database.DB.Query(`
        SELECT id, moderator_id, post_id, comment_id, reason, created_at, status
        FROM reports
        WHERE status = 'pending'
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []models.Report
	for rows.Next() {
		var r models.Report
		var postID, commentID *int
		err := rows.Scan(&r.ID, &r.ModeratorID, &postID, &commentID, &r.Reason, &r.CreatedAt, &r.Status)
		if err != nil {
			return nil, err
		}
		r.PostID = postID
		r.CommentID = commentID
		reports = append(reports, r)
	}
	return reports, nil
}

func UpdateReportStatus(reportID int, status string) error {
	_, err := database.DB.Exec("UPDATE reports SET status = ? WHERE id = ?", status, reportID)
	return err
}
