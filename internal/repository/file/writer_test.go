package file_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gustavo-silva98/adnotes/internal/repository/file"
)

func setupTestDB(t *testing.T) *file.SqliteHandler {
	t.Helper()
	db_path := ":memory:"
	ctx := context.Background()

	handler, err := file.InitDB(db_path, ctx)
	if err != nil {
		t.Fatalf("Erro ao inicializar banco de teste - %v", err)
	}
	return handler
}

func TestInsertNote(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	note := &file.Note{
		Hour:         int64(1),
		NoteText:     "Teste de insert",
		Reminder:     1,
		PlusReminder: 2,
	}

	id, err := handler.InsertNote(note, ctx)
	if err != nil {
		t.Fatalf("Falha ao inserir nota teste - %v", err)
	}
	if id <= 0 {
		t.Errorf("ID de insert Note retornado inválido - %d", id)
	}
}

func TestQueryNote(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	note := &file.Note{
		Hour:         int64(1),
		NoteText:     "Teste de insert",
		Reminder:     1,
		PlusReminder: 2,
	}

	_, _ = handler.InsertNote(note, ctx)

	result, err := handler.QueryNote(10, 0, ctx)
	if err != nil {
		t.Fatalf("Erro ao consultar notas - %v", err)
	}

	if len(result) == 0 {
		t.Errorf("Nenhuma nota retornada")
	}
}

func TestUpdateEditNote(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	note := &file.Note{
		Hour:         int64(1),
		NoteText:     "Teste de insert",
		Reminder:     1,
		PlusReminder: 2,
	}

	id, _ := handler.InsertNote(note, ctx)
	note.ID = int(id)
	note.NoteText = "Nota atualizada"

	rowsAffected, err := handler.UpdateEditNoteRepository(ctx, *note)
	if err != nil {
		t.Fatalf("Erro ao atualizar nota - %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Esperado 1 linha afetada. Resultado divergente")
	}

}

func TestDeleteNoteRepository(t *testing.T) {
	ctx := context.Background()
	handler, _ := file.InitDB(":memory:", ctx)
	note := &file.Note{
		Hour:         int64(1),
		NoteText:     "Teste de insert",
		Reminder:     1,
		PlusReminder: 2,
	}
	id, _ := handler.InsertNote(note, ctx)

	rowsAffected, err := handler.DeleteNoteRepository(ctx, int(id))
	if err != nil {
		t.Fatalf("Falha ao deletar nota no teste - %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Era esperado somente 1 linha. Valor divergente.")
	}
}

func BenchmarkInsertNote(b *testing.B) {
	ctx := context.Background()
	handler, _ := file.InitDB(":memory:", ctx)

	for i := 0; i < b.N; i++ {
		note := &file.Note{
			Hour:         int64(i),
			NoteText:     "Benchmark",
			Reminder:     1,
			PlusReminder: 2,
		}
		_, _ = handler.InsertNote(note, ctx)
	}

}

func FTSTableExists(db *file.SqliteHandler) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' AND name='notes_fts'`

	row := db.DB.QueryRow(query)
	var name string
	err := row.Scan(&name)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func TestCreateFTSTable(t *testing.T) {
	db, err := sql.Open("sqlite", "teste_db.db")
	if err != nil {
		t.Errorf("Erro ao criar o BD teste - Err: %v", err)
	}

	handler := &file.SqliteHandler{
		DbPath:    "teste_db.db",
		TableName: "teste_notas",
		DB:        db,
	}

	err = handler.CreateFTSTable()
	if err != nil {
		db.Close()
		os.Remove("teste_db.db")
		log.Fatalf("Erro ao criar tabela FTS: %v", err)
	}
	exists, err := FTSTableExists(handler)
	if err != nil {
		db.Close()
		os.Remove("teste_db.db")
		log.Fatalf("Erro ao verificar tabela FTS: %v", err)
	}

	if exists {
		log.Println("Tabela FTS Criada com sucesso")
	} else {
		log.Println("Tabela FTS não foi criada")
	}
	db.Close()
	err = os.Remove("teste_db.db")
	if err != nil {
		t.Errorf("Erro ao remover arquivos de teste: %v", err)
	}

}

func TriggerExists(db *file.SqliteHandler, triggerName string) (bool, error) {
	query := `SELECT name FROM sqlite_master WHERE type='trigger' AND name=?`

	row := db.DB.QueryRow(query, triggerName)
	var name string
	err := row.Scan(&name)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil

}
func AllTriggersExist(db *file.SqliteHandler) (bool, error) {
	triggers := []string{"notes_ai", "notes_ad", "notes_au"}

	for _, trigger := range triggers {
		exists, err := TriggerExists(db, trigger)
		if err != nil {
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func TestCreateFTSTroggers(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	err := handler.CreateFTSTriggers(ctx)
	if err != nil {
		t.Fatalf("Erro ao criar triggers FTS: %v", err)
	}
	exists, err := AllTriggersExist(handler)
	if err != nil {
		t.Errorf("Erro ao verificar triggers: %v", err)
	}

	if !exists {
		t.Errorf("Nem todos os triggers foram criados")
	}
}

func TestFTSTriggersIntegration(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	// Testar trigger de INSERT
	note := &file.Note{
		Hour:         123456789,
		NoteText:     "Teste trigger insert",
		Reminder:     1,
		PlusReminder: 2,
	}

	id, err := handler.InsertNote(note, ctx)
	if err != nil {
		t.Fatalf("Erro ao inserir nota: %v", err)
	}

	var ftsText string
	err = handler.DB.QueryRowContext(ctx,
		"SELECT note_text_fts FROM notes_fts WHERE rowid = ?", id).Scan(&ftsText)
	if err != nil {
		t.Fatalf("Trigger de INSERT não funcionou: %v", err)
	}

	if ftsText != note.NoteText {
		t.Errorf("Texto no FTS não confere. Esperado: %s, Obtido: %s",
			note.NoteText, ftsText)
	}

	_, err = handler.DeleteNoteRepository(ctx, int(id))
	if err != nil {
		t.Fatalf("Erro ao deletar nota: %v", err)
	}

	var count int
	err = handler.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notes_fts WHERE rowid = ?", id).Scan(&count)

	if err != nil {
		t.Fatalf("Erro ao verificar exclusão no FTS: %v", err)
	}
	if count != 0 {
		t.Errorf("Trigger de DELETE não funcionou. Linhas restantes %v", count)
	}
}

func TestFullSearchNote(t *testing.T) {
	handler := setupTestDB(t)
	ctx := context.Background()

	// Testar trigger de INSERT
	note := &file.Note{
		Hour:         123456789,
		NoteText:     "Teste trigger insert",
		Reminder:     1,
		PlusReminder: 2,
	}

	id, err := handler.InsertNote(note, ctx)
	if err != nil {
		t.Fatalf("Erro ao inserir nota: %v", err)
	}
	results, err := handler.FullSearchNote(ctx, "insert")
	if err != nil {
		t.Error("Falha ao realizar busca de notas")
	}
	val, ok := results[int(id)]
	if !ok {
		t.Error("Id não foi retornado ao realizar a Full Search")
	}
	if val.NoteText != note.NoteText {
		t.Error("Texto da nota retornado pela busca está divergente do esperado")
	}

	fmt.Println(results[int(id)])
}
