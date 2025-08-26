package utils

// CreatePost inserts a new post into the database.
func (db *DataBase) CreatePost(authorUUID, title, content string) error {
	_, err := db.Conn.Exec("INSERT INTO posts (title, content, author_uuid) VALUES (?, ?, ?)", title, content, authorUUID)
	return err
}

// AddComment adds a new comment to a post.
func (db *DataBase) AddComment(authorUUID string, postID int, content string) error {
	_, err := db.Conn.Exec("INSERT INTO comments (content, comment_author_uuid, post_id) VALUES (?, ?, ?)", content, authorUUID, postID)
	return err
}

// ToggleLike records a like or dislike for a post from a user.
// If like is true, it records a like; otherwise a dislike.
// Existing interactions from the same user are removed before insertion.
func (db *DataBase) ToggleLike(userUUID string, postID int, like bool) error {
	db.Write.Lock()
	defer db.Write.Unlock()

	_, err := db.Conn.Exec("DELETE FROM interactions WHERE user_uuid = ? AND post_id = ?", userUUID, postID)
	if err != nil {
		return err
	}

	var liked, disliked bool
	if like {
		liked = true
	} else {
		disliked = true
	}

	_, err = db.Conn.Exec("INSERT INTO interactions (user_uuid, post_id, liked, disliked) VALUES (?, ?, ?, ?)", userUUID, postID, liked, disliked)
	return err
}

// GetPosts retrieves all posts with their comments and interactions.
func (db *DataBase) GetPosts() ([]Post, error) {
	rows, err := db.Conn.Query("SELECT p.id, p.title, p.content, u.uuid, u.username FROM posts p JOIN users u ON p.author_uuid = u.uuid ORDER BY p.id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var auuid, ausername string
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &auuid, &ausername); err != nil {
			return nil, err
		}
		p.Author = User{UUID: auuid, Username: ausername}

		// Fetch comments for the post
		cRows, err := db.Conn.Query("SELECT c.id, c.content, u.uuid, u.username FROM comments c JOIN users u ON c.comment_author_uuid = u.uuid WHERE c.post_id = ?", p.ID)
		if err == nil {
			for cRows.Next() {
				var c Comment
				var cuuid, cusername string
				if err := cRows.Scan(&c.ID, &c.Content, &cuuid, &cusername); err == nil {
					c.Author = User{UUID: cuuid, Username: cusername}
					p.Comments = append(p.Comments, c)
				}
			}
			cRows.Close()
		}

		// Fetch likes and dislikes
		iRows, err := db.Conn.Query("SELECT user_uuid, liked, disliked FROM interactions WHERE post_id = ?", p.ID)
		if err == nil {
			for iRows.Next() {
				var inter Interaction
				if err := iRows.Scan(&inter.User.UUID, &inter.Like, &inter.DisLike); err == nil {
					if inter.Like {
						p.Likes = append(p.Likes, inter)
					} else if inter.DisLike {
						p.DisLikes = append(p.DisLikes, inter)
					}
				}
			}
			iRows.Close()
		}

		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
