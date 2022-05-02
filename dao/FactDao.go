package dao

import (
	"log"
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"mhbotgo.com/domain"
	"mhbotgo.com/profiler"
)

// FactDao : Contains data access methods for stored user fact
type FactDao struct {
	Session *discordgo.Session
}

// GetFactByUser : Gets a fact by a user ID
func (d *FactDao) GetFactByUser(u *discordgo.User) (*domain.Fact, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from UserFacts where UserID = ? limit 1"
	stmt, err := queryForRows(query, DB, u.ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getFactFromStmt(stmt, d)
}

// GetRandomFact : Gets a random fact
func (d *FactDao) GetRandomFact() (*domain.Fact, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from UserFacts order by RANDOM() limit 1"
	stmt, err := queryForRows(query, DB)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getFactFromStmt(stmt, d)
}

// InsertFact : Insert a new fact
func (d *FactDao) InsertFact(fact *domain.Fact, s *discordgo.Session) *domain.Fact {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `INSERT INTO UserFacts 
	(UserID, FactContent, LastUsedUnix) 
	VALUES
	(?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Println("Error inserting server", err)

		return nil
	}
	defer stmt.Close()

	// Default 0 unix time for last used time
	err = stmt.Exec(fact.UserID, fact.FactContent, 0)
	if err != nil {
		log.Println("Error inserting fact", err)
		return nil
	}

	return fact
}

// UpdateFactByUser : Updates a fact by object
// Returns user ID of fact
func (d *FactDao) UpdateFactByUser(fact *domain.Fact, u *discordgo.User) string {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "UPDATE UserFacts " +
		"	SET " +
		"		UserID = ?," +
		"		FactContent = ?," +
		"		LastUsedUnix = ?" +
		"	WHERE " +
		"		UserID = ?"

	stmt, err := DB.Prepare(query,
		fact.UserID, fact.FactContent, fact.LastUsedUnix)
	if err != nil {
		log.Println("Error updating fact", err)
		return "-1"
	}
	defer stmt.Close()

	stmt.Exec()

	return fact.UserID
}

// DeleteFactByUserID : Deletes a fact by given user ID
// Returns ID of deleted fact
func (d *FactDao) DeleteFactByUserID(discordID string) bool {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `	DELETE FROM UserFacts  
				WHERE UserID = ?`

	stmt, err := DB.Prepare(query, discordID)
	if err != nil {
		return false
	}

	defer stmt.Close()

	stmt.Exec()
	return true
}

func mapRowToFact(rows *sqlite3.Stmt, s *discordgo.Session) (*domain.Fact, error) {
	fact := new(domain.Fact)

	err := rows.Scan(
		&fact.UserID,
		&fact.FactContent,
		&fact.LastUsedUnix)

	if err != nil {
		return nil, err
	}
	profiler.Start()

	return fact, err
}

func getFactFromStmt(stmt *sqlite3.Stmt, d *FactDao) (*domain.Fact, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}

	fact, err := mapRowToFact(stmt, d.Session)
	if err != nil {
		return nil, err
	}

	fact, err = mapFactORMFields(fact, d.Session)
	if err != nil {
		return nil, err
	}
	return fact, err
}

func mapFactORMFields(fact *domain.Fact, s *discordgo.Session) (*domain.Fact, error) {

	fact.User, _ = s.User(fact.UserID)
	fact.LastUsedTime = time.Unix(fact.LastUsedUnix, 0)

	return fact, nil
}

func processStmtForFactArray(stmt *sqlite3.Stmt, d *FactDao) ([]*domain.Fact, error) {
	facts := make([]*domain.Fact, 0)
	for {
		fact, err := getFactFromStmt(stmt, d)
		if err != nil {
			return nil, err
		}
		if fact != nil {
			facts = append(facts, fact)
		} else if len(facts) > 0 {
			break
		} else {
			return nil, nil // errors.New("Couldn't find event")
		}
	}
	return facts, nil
}

func processStmtForFact(stmt *sqlite3.Stmt, d *FactDao) (*domain.Fact, error) {
	fact, err := getFactFromStmt(stmt, d)
	if err != nil {
		stmt.Close()
		return nil, err
	}
	return fact, err
}
