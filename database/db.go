package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	"forum/models"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal("Erreur ouverture DB:", err)
	}

	// Active les clés étrangères (désactivées par défaut dans SQLite)
	DB.Exec("PRAGMA foreign_keys = ON")

	rows, _ := DB.Query("PRAGMA table_info(users)")
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull int
		var dflt interface{}
		var pk int
		rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
		log.Println("colonne:", name)
	}

	// Lit et exécute le schéma
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		log.Fatal("Impossible de lire schema.sql:", err)
	}
	if _, err = DB.Exec(string(schema)); err != nil {
		log.Fatal("Erreur exécution schéma:", err)
	}

	log.Println("Base de données initialisée")

}

// ─── USERS ───────────────────────────────────────────────────────────────────

func CreateUser(username, email, hash string) error {
	_, err := DB.Exec(
		`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		username, email, hash,
	)
	return err
}

func GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := DB.QueryRow(
		`SELECT id, username, email, password_hash, avatar_url, bio, created_at
		 FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.Bio, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(id int) (*models.User, error) {
	u := &models.User{}
	err := DB.QueryRow(
		`SELECT id, username, email, password_hash, avatar_url, bio, created_at
		 FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.Bio, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ─── SESSIONS ────────────────────────────────────────────────────────────────

func CreateSession(uuid string, userID int, expires time.Time) error {
	_, err := DB.Exec(
		`INSERT INTO sessions (uuid, user_id, expires_at) VALUES (?, ?, ?)`,
		uuid, userID, expires,
	)
	return err
}

func GetSession(uuid string) (*models.Session, error) {
	s := &models.Session{}
	err := DB.QueryRow(
		`SELECT uuid, user_id, expires_at FROM sessions WHERE uuid = ? AND expires_at > CURRENT_TIMESTAMP`,
		uuid,
	).Scan(&s.UUID, &s.UserID, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func DeleteSession(uuid string) error {
	_, err := DB.Exec(`DELETE FROM sessions WHERE uuid = ?`, uuid)
	return err
}

func CleanExpiredSessions() {
	DB.Exec(`DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP`)
}

// ─── CATEGORIES ──────────────────────────────────────────────────────────────

func GetCategories() ([]models.Category, error) {
	rows, err := DB.Query(`SELECT id, name, slug FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []models.Category
	for rows.Next() {
		var c models.Category
		rows.Scan(&c.ID, &c.Name, &c.Slug)
		cats = append(cats, c)
	}
	return cats, nil
}

func GetCategoryBySlug(slug string) (*models.Category, error) {
	c := &models.Category{}
	err := DB.QueryRow(
		`SELECT id, name, slug FROM categories WHERE slug = ?`, slug,
	).Scan(&c.ID, &c.Name, &c.Slug)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ─── TOPICS ──────────────────────────────────────────────────────────────────

func CreateTopic(userID, categoryID int, title, description string) (int64, error) {
	res, err := DB.Exec(
		`INSERT INTO topics (user_id, category_id, title, description) VALUES (?, ?, ?, ?)`,
		userID, categoryID, title, description,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTopics retourne tous les topics avec username, catégorie, nb posts et nb likes
func GetTopics(filter string, userID int) ([]models.Topic, error) {
	query := `
		SELECT
			t.id, t.user_id, t.category_id, t.title, t.description, t.created_at,
			u.username,
			c.name, c.slug,
			COUNT(DISTINCT p.id) AS post_count,
			COUNT(DISTINCT r.post_id) AS like_count
		FROM topics t
		JOIN users u      ON u.id = t.user_id
		JOIN categories c ON c.id = t.category_id
		LEFT JOIN posts p ON p.topic_id = t.id
		LEFT JOIN reactions r ON r.post_id = p.id AND r.type = 'like'`

	var args []interface{}

	switch filter {
	case "mine":
		query += ` WHERE t.user_id = ?`
		args = append(args, userID)
	case "liked":
		query += `
		WHERE EXISTS (
			SELECT 1 FROM posts p2
			JOIN reactions r2 ON r2.post_id = p2.id
			WHERE p2.topic_id = t.id AND r2.user_id = ? AND r2.type = 'like'
		)`
		args = append(args, userID)
	}

	query += ` GROUP BY t.id ORDER BY t.created_at DESC`

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		rows.Scan(
			&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.Description, &t.CreatedAt,
			&t.Username, &t.CategoryName, &t.CategorySlug,
			&t.PostCount, &t.LikeCount,
		)
		topics = append(topics, t)
	}
	return topics, nil
}

func GetTopicsByCategory(slug string) ([]models.Topic, error) {
	rows, err := DB.Query(`
		SELECT
			t.id, t.user_id, t.category_id, t.title, t.description, t.created_at,
			u.username, c.name, c.slug,
			COUNT(DISTINCT p.id), COUNT(DISTINCT r.post_id)
		FROM topics t
		JOIN users u      ON u.id = t.user_id
		JOIN categories c ON c.id = t.category_id
		LEFT JOIN posts p ON p.topic_id = t.id
		LEFT JOIN reactions r ON r.post_id = p.id AND r.type = 'like'
		WHERE c.slug = ?
		GROUP BY t.id ORDER BY t.created_at DESC`, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		rows.Scan(
			&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.Description, &t.CreatedAt,
			&t.Username, &t.CategoryName, &t.CategorySlug,
			&t.PostCount, &t.LikeCount,
		)
		topics = append(topics, t)
	}
	return topics, nil
}

func GetTopicByID(id int) (*models.Topic, error) {
	t := &models.Topic{}
	err := DB.QueryRow(`
        SELECT t.id, t.user_id, t.category_id, t.title, t.description, t.created_at,
               u.username, c.name, c.slug
        FROM topics t
        JOIN users u      ON u.id = t.user_id
        JOIN categories c ON c.id = t.category_id
        WHERE t.id = ?`, id,
	).Scan(
		&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.Description, &t.CreatedAt,
		&t.Username, &t.CategoryName, &t.CategorySlug,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// ─── POSTS ───────────────────────────────────────────────────────────────────

func CreatePost(topicID, userID int, content, imageURL string, parentID *int) error {
	_, err := DB.Exec(
		`INSERT INTO posts (topic_id, user_id, content, image_url, parent_id) VALUES (?, ?, ?, ?, ?)`,
		topicID, userID, content, imageURL, parentID,
	)
	return err
}

func GetPostsByTopic(topicID, currentUserID int) ([]models.Post, error) {
	rows, err := DB.Query(`
		SELECT
			p.id, p.topic_id, p.user_id, p.content, 
			COALESCE(p.image_url, ''),
			p.parent_id,
			-- MODIFICATION ICI : On concatène [image] si parent.image_url n'est pas vide
			COALESCE(parent.content, '') || CASE 
				WHEN parent.image_url IS NOT NULL AND parent.image_url != '' THEN ' [image]' 
				ELSE '' 
			END,
			COALESCE(parent_user.username, ''),
			p.created_at,
			u.username, u.avatar_url,
			SUM(CASE WHEN r.type = 'like'    THEN 1 ELSE 0 END) AS likes,
			SUM(CASE WHEN r.type = 'dislike' THEN 1 ELSE 0 END) AS dislikes,
			COALESCE((SELECT type FROM reactions WHERE user_id = ? AND post_id = p.id), '') AS user_reaction
		FROM posts p
		JOIN users u ON u.id = p.user_id
		LEFT JOIN posts parent ON parent.id = p.parent_id
		LEFT JOIN users parent_user ON parent_user.id = parent.user_id
		LEFT JOIN reactions r ON r.post_id = p.id
		WHERE p.topic_id = ?
		GROUP BY p.id
		ORDER BY p.created_at ASC`, currentUserID, topicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		rows.Scan(
			&p.ID, &p.TopicID, &p.UserID, &p.Content,
			&p.ImageURL, &p.ParentID, &p.ParentContent, &p.ParentAuthor,
			&p.CreatedAt, &p.Username, &p.AvatarURL,
			&p.Likes, &p.Dislikes, &p.UserReaction,
		)
		posts = append(posts, p)
	}
	return posts, nil
}

// ─── REACTIONS ───────────────────────────────────────────────────────────────

// ToggleReaction : si la même réaction existe déjà → supprime (toggle off), sinon insère/remplace
func ToggleReaction(userID, postID int, reactionType string) error {
	var existing string
	err := DB.QueryRow(
		`SELECT type FROM reactions WHERE user_id = ? AND post_id = ?`,
		userID, postID,
	).Scan(&existing)

	if err == nil && existing == reactionType {
		// même réaction → on annule
		_, err = DB.Exec(`DELETE FROM reactions WHERE user_id = ? AND post_id = ?`, userID, postID)
		return err
	}

	// nouvelle réaction ou changement like↔dislike
	_, err = DB.Exec(
		`INSERT OR REPLACE INTO reactions (user_id, post_id, type) VALUES (?, ?, ?)`,
		userID, postID, reactionType,
	)
	return err
}

// ─── RECHERCHE ───────────────────────────────────────────────────────────────
func SearchTopics(query string) ([]models.Topic, error) {
	q := "%" + query + "%"
	rows, err := DB.Query(`
		SELECT
			t.id, t.user_id, t.category_id, t.title, t.created_at,
			u.username, c.name, c.slug,
			COUNT(DISTINCT p.id), COUNT(DISTINCT r.post_id)
		FROM topics t
		JOIN users u      ON u.id = t.user_id
		JOIN categories c ON c.id = t.category_id
		LEFT JOIN posts p ON p.topic_id = t.id
		LEFT JOIN reactions r ON r.post_id = p.id AND r.type = 'like'
		WHERE t.title LIKE ? OR u.username LIKE ?
		GROUP BY t.id ORDER BY t.created_at DESC`, q, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []models.Topic
	for rows.Next() {
		var t models.Topic
		rows.Scan(
			&t.ID, &t.UserID, &t.CategoryID, &t.Title, &t.CreatedAt,
			&t.Username, &t.CategoryName, &t.CategorySlug,
			&t.PostCount, &t.LikeCount,
		)
		topics = append(topics, t)
	}
	return topics, nil
}

func SearchUsernames(q string) ([]string, error) {
	rows, err := DB.Query(
		`SELECT username FROM users WHERE username LIKE ? LIMIT 8`,
		q+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var u string
		rows.Scan(&u)
		usernames = append(usernames, u)
	}
	return usernames, nil
}

func GetUserByUsername(username string) (*models.User, error) {
	u := &models.User{}
	err := DB.QueryRow(
		`SELECT id, username, email, password_hash, avatar_url, bio, created_at
		 FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarURL, &u.Bio, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUserProfile(userID int, bio, avatarURL string) error {
	_, err := DB.Exec(
		`UPDATE users SET bio = ?, avatar_url = ? WHERE id = ?`,
		bio, avatarURL, userID,
	)
	return err
}

func GetFavoriteTopics(userID int) ([]models.FavoriteTopic, error) {
	rows, err := DB.Query(`
        SELECT t.id, t.title, COUNT(p.id) as post_count
        FROM topics t
        LEFT JOIN posts p ON p.topic_id = t.id
        WHERE EXISTS (
            SELECT 1 FROM posts p2
            JOIN reactions r ON r.post_id = p2.id
            WHERE p2.topic_id = t.id AND r.user_id = ? AND r.type = 'like'
        ) OR t.user_id = ?
        GROUP BY t.id
        ORDER BY post_count DESC, t.created_at DESC
        LIMIT 10`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favs []models.FavoriteTopic
	for rows.Next() {
		var f models.FavoriteTopic
		rows.Scan(&f.TopicID, &f.TopicTitle, &f.PostCount)
		favs = append(favs, f)
	}
	return favs, nil
}
