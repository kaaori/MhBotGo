package dao

import (
	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/profiler"
)

// BirthdayDao : Contains data access methods for stored user getFactFromStmt
type FactDao struct {
	Session *discordgo.Session
}

// GetFactByUser : Gets a fact by a user ID
func (d *FactDao) GetFactByUser(g *discordgo.Guild, u *discordgo.User) (*domain.Fact, error) {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := "select * from UserFacts where UserID = ? limit 1"
	stmt, err := queryForRows(query, DB, u.ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getFactFromStmt(stmt, d, g)
}

// InsertFact : Insert a new fact
func (d *FactDao) InsertFact(fact *domain.Fact, s *discordgo.Session, g *discordgo.Guild) *domain.Fact {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `INSERT INTO UserFacts 
	(UserID, FactContent, LastUsedUnix) 
	VALUES
	(?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Error("Error inserting server", err)

		return nil
	}
	defer stmt.Close()

	err = stmt.Exec(fact.UserID, fact.FactContent, fact.LastUsedUnix)
	if err != nil {
		log.Error("Error inserting fact", err)
		return nil
	}

	return fact
}

// UpdateBirthdayByUser : Updates a fact by object
// Returns user ID of fact
func (d *FactDao) UpdateBirthdayByUser(fact *domain.Fact, u *discordgo.User) string {
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
		log.Error("Error updating fact", err)
		return "-1"
	}
	defer stmt.Close()

	stmt.Exec()

	return fact.UserID
}

// DeleteFactByUserID : Deletes a fact by given user ID
// Returns ID of deleted fact
func (d *FactDao) DeleteFactByUserID(discordID string) int64 {
	DB := GetConnection(ConnString)
	defer DB.Close()

	query := `	DELETE FROM UserFacts  
				WHERE UserID = ?`

	stmt, err := DB.Prepare(query, discordID)
	if err != nil {
		return -1
	}

	defer stmt.Close()

	stmt.Exec()
	return DB.LastInsertRowID()
}

func mapRowToFact(rows *sqlite3.Stmt, s *discordgo.Session, g *discordgo.Guild) (*domain.Fact, error) {
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

func getFactFromStmt(stmt *sqlite3.Stmt, d *FactDao, g *discordgo.Guild) (*domain.Fact, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}
	fact, err := mapRowToFact(stmt, d.Session, g)
	if err != nil {
		return nil, err
	}
	return fact, err
}

func processStmtForFactArray(stmt *sqlite3.Stmt, d *FactDao, g *discordgo.Guild) ([]*domain.Fact, error) {
	facts := make([]*domain.Fact, 0)
	for {
		fact, err := getFactFromStmt(stmt, d, g)
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

func processStmtForFact(stmt *sqlite3.Stmt, d *FactDao, g *discordgo.Guild) (*domain.Fact, error) {
	birthday, err := getFactFromStmt(stmt, d, g)
	if err != nil {
		stmt.Close()
		return nil, err
	}
	return birthday, err
}
