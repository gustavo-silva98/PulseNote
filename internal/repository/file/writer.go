package file

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Writer interface {
	InsertNote(n *Note, ctx context.Context) (int64, error)
	QueryNote(limit int, offset int, ctx context.Context) (map[int]Note, error)
	UpdateEditNoteRepository(ctx context.Context, note Note) (int64, error)
	DeleteNoteRepository(ctx context.Context, noteId int) (int64, error)
	FullSearchNote(ctx context.Context, argQuery string) (map[int]Note, error)
}

type SqliteHandler struct {
	DbPath    string
	TableName string
	DB        *sql.DB
}

type Note struct {
	ID           int
	Hour         int64
	NoteText     string
	Reminder     int
	PlusReminder int
}

func InitDB(pathString string, ctx context.Context) (*SqliteHandler, error) {
	db, err := sql.Open("sqlite", pathString)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS notas (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hour INTEGER NOT NULL,
			note_text TEXT NOT NULL,
			reminder INTEGER,
			plusreminder INTEGER
		)`,
	)
	if err != nil {
		return nil, err
	}
	sql_db := &SqliteHandler{
		DbPath:    pathString,
		TableName: "notas",
		DB:        db,
	}
	err = sql_db.CreateFTSTable()
	if err != nil {
		return nil, err
	}
	err = sql_db.CreateFTSTriggers(ctx)
	if err != nil {
		return nil, err
	}

	return sql_db, nil
}

func (s SqliteHandler) InsertNote(n *Note, ctx context.Context) (int64, error) {

	res, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO notas (hour, note_text, reminder, plusreminder) VALUES (?, ?, ?, ?)`,
		n.Hour, n.NoteText, n.Reminder, n.PlusReminder,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s SqliteHandler) QueryNote(limit int, offset int, ctx context.Context) (map[int]Note, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT * FROM notas ORDER BY id DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var queryMap = map[int]Note{}
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Hour, &note.NoteText, &note.Reminder, &note.PlusReminder)
		if err != nil {
			return nil, err
		}
		queryMap[note.ID] = note
	}

	return queryMap, nil

}

func WriteLog(msg string, logFilePath string) {
	if logFilePath == "" {
		logFilePath = filepath.Join("..", "/data/banco.db")
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	log.SetOutput(file)

	log.SetFlags(log.LstdFlags)

	log.Println(msg)
}

func (s SqliteHandler) GetFirsIndexPage(ctx context.Context) (int, error) {
	row, err := s.DB.QueryContext(
		ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM %v`, s.TableName),
	)

	if err != nil {
		return 0, err
	}
	defer row.Close()
	var count int
	for row.Next() {
		if err := row.Scan(&count); err != nil {
			return 0, nil
		}
	}
	if count < 10 {
		return 10, nil
	}
	return count, nil
}

func (s SqliteHandler) UpdateEditNoteRepository(ctx context.Context, note Note) (int64, error) {
	row, err := s.DB.ExecContext(
		ctx,
		`UPDATE notas
		SET hour = ?, note_text = ?, reminder = ?, plusreminder = ?
		WHERE id = ?`,
		note.Hour, note.NoteText, note.Reminder, note.PlusReminder, note.ID)
	if err != nil {
		return 0, err
	}
	ra, err := row.RowsAffected()
	if err != nil {
		return 0, err
	}

	return ra, nil

}

func (s SqliteHandler) DeleteNoteRepository(ctx context.Context, noteId int) (int64, error) {

	row, err := s.DB.ExecContext(ctx, `DELETE FROM notas WHERE id = ?`, noteId)
	if err != nil {
		return 0, err
	}
	ra, err := row.RowsAffected()
	if err != nil {
		return 0, err
	}

	return ra, nil
}

func (s SqliteHandler) CreateFTSTable() error {
	createFTSQuery := `CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(note_text_fts);`

	_, err := s.DB.Exec(createFTSQuery)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela FTS: %v", err)
	}
	return nil
}

func (s SqliteHandler) CreateFTSTriggers(ctx context.Context) error {
	triggerInsert := `
	CREATE TRIGGER IF NOT EXISTS notes_ai AFTER INSERT ON notas BEGIN 
		INSERT INTO notes_fts(rowid,note_text_fts) VALUES (new.id, new.note_text);
	END;`

	triggerDelete := `
	CREATE TRIGGER IF NOT EXISTS notes_ad AFTER DELETE ON notas BEGIN 
		DELETE FROM notes_fts WHERE rowid = old.id;
	END;`

	triggerUpdate := `
	CREATE TRIGGER IF NOT EXISTS notes_au AFTER UPDATE ON notas BEGIN 
		DELETE FROM notes_fts WHERE rowid = old.id;
		INSERT INTO notes_fts(rowid,note_text_fts) VALUES (new.id, new.note_text);
	END;`

	if _, err := s.DB.ExecContext(ctx, triggerInsert); err != nil {
		return fmt.Errorf("erro ao criar trigger de INSERT: %v", err)
	}

	if _, err := s.DB.ExecContext(ctx, triggerDelete); err != nil {
		return fmt.Errorf("erro ao criar trigger de DELETE: %v", err)
	}

	if _, err := s.DB.ExecContext(ctx, triggerUpdate); err != nil {
		return fmt.Errorf("erro ao criar trigger de UPDATE: %v", err)
	}
	return nil
}

func (s SqliteHandler) FullSearchNote(ctx context.Context, argQuery string) (map[int]Note, error) {
	if argQuery != "" {
		argQuery = argQuery + "*"
	}

	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT nt.id, nt.hour, fts.note_text_fts, nt.reminder, nt.plusreminder
			FROM notes_fts fts
			INNER JOIN notas nt ON nt.id = fts.rowid
			WHERE fts.note_text_fts MATCH ?`, argQuery,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queryMap = map[int]Note{}
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Hour, &note.NoteText, &note.Reminder, &note.PlusReminder)
		if err != nil {
			return nil, err
		}
		queryMap[note.ID] = note
	}

	return queryMap, nil
}
