package dao

import (
	"time"

	"github.com/bvinc/go-sqlite-lite/sqlite3"
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
	"github.com/kaaori/MhBotGo/profiler"
)

// BirthdayDao : Contains data access methods for stored user birthdays
type BirthdayDao struct {
	Session *discordgo.Session
}

// GetAllBirthdaysForServerForWeek : Gets a server's user's birthdays within a week range
func (d *BirthdayDao) GetAllBirthdaysForServerForWeek(serverID string, weekTime time.Time, g *discordgo.Guild) ([]*domain.Birthday, error) {
	DB := GetConnection()
	defer DB.Close()

	query := "select * from Birthdays where ServerID = ? and BirthMonthNum = ? and BirthDayNum between ? AND ?"

	profiler.Start()
	stmt, err := queryForRows(query, DB, serverID, int(weekTime.Month()), weekTime.Day(), weekTime.AddDate(0, 0, 7).Day())
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	profiler.StopAndPrintSeconds("Querying for rows")

	return processStmtForBirthdayArray(stmt, d, g)
}

// GetBirthdayByID : Gets a birthday by its ID
func (d *BirthdayDao) GetBirthdayByID(ID int64, g *discordgo.Guild, u *discordgo.User) (*domain.Birthday, error) {
	DB := GetConnection()
	defer DB.Close()

	query := "select * from Birthdays where BirthdayID = ?"
	stmt, err := queryForRows(query, DB, ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getBirthdayFromStmt(stmt, d, g)
}

// GetBirthdayByUser : Gets a birthday by a user ID
func (d *BirthdayDao) GetBirthdayByUser(g *discordgo.Guild, u *discordgo.User) (*domain.Birthday, error) {
	DB := GetConnection()
	defer DB.Close()

	query := "select * from Birthdays where UserID = ? limit 1"
	stmt, err := queryForRows(query, DB, u.ID)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	return getBirthdayFromStmt(stmt, d, g)
}

// InsertBirthday : Insert a new birthday
func (d *BirthdayDao) InsertBirthday(birthday *domain.Birthday, s *discordgo.Session, g *discordgo.Guild) *domain.Birthday {
	DB := GetConnection()
	defer DB.Close()

	query := `INSERT INTO Birthdays 
	(ServerID, UserID, BirthMonthNum, BirthDayNum) 
	VALUES
	(?,?,?,?)`

	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Error("Error inserting server", err)

		return nil
	}
	defer stmt.Close()

	// We need to increment the birthday by 1 due to how it's being parsed
	err = stmt.Exec(birthday.ServerID, birthday.GuildUserID, birthday.BirthdayMonth, birthday.BirthdayDay)
	if err != nil {
		log.Error("Error inserting birthday", err)
		return nil
	}

	birthday, err = mapBirthdayORMFields(birthday, s, g)
	if err != nil {
		log.Error("Error mapping ORM fields in new birthday", err)
		return nil
	}
	birthday.BirthdayID = DB.LastInsertRowID()

	return birthday
}

// UpdateBirthdayByUser : Updates a birthday by object
// Returns ID of new birthday
func (d *BirthdayDao) UpdateBirthdayByUser(birthday *domain.Birthday, u *discordgo.User) int64 {
	DB := GetConnection()
	defer DB.Close()

	query := "UPDATE Birthdays " +
		"	SET " +
		"		ServerID = ?," +
		"		UserID = ?," +
		"		BirthMonthNum = ?," +
		"		BirthDayNum = ?" +
		"	WHERE " +
		"		UserID = ?"

	stmt, err := DB.Prepare(query,
		birthday.ServerID, birthday.GuildUserID, birthday.BirthdayMonth, birthday.BirthdayDay, u.ID)
	if err != nil {
		log.Error("Error updating birthday", err)
		return -1
	}
	defer stmt.Close()

	stmt.Exec()

	return birthday.BirthdayID
}

// DeleteBirthdayByUserID : Deletes a birthday by given user ID
// Returns ID of deleted event
func (d *BirthdayDao) DeleteBirthdayByUserID(discordID string) int64 {
	DB := GetConnection()
	defer DB.Close()

	query := `	DELETE FROM Events  
				WHERE UserID = ?`

	stmt, err := DB.Prepare(query, discordID)
	if err != nil {
		return -1
	}

	defer stmt.Close()

	stmt.Exec()
	return DB.LastInsertRowID()
}

func mapRowToBirthday(rows *sqlite3.Stmt, s *discordgo.Session, g *discordgo.Guild) (*domain.Birthday, error) {
	birthday := new(domain.Birthday)

	err := rows.Scan(
		&birthday.BirthdayID,
		&birthday.ServerID,
		&birthday.GuildUserID,
		&birthday.BirthdayMonth,
		&birthday.BirthdayDay)

	if err != nil {
		return nil, err
	}
	profiler.Start()

	birthday, err = mapBirthdayORMFields(birthday, s, g)
	if err != nil {
		log.Error("Error mapping ORM Fields in event", err)
		return nil, err
	}
	profiler.StopAndPrintSeconds("Mapping ORM fields")
	return birthday, err
}

func mapBirthdayORMFields(birthday *domain.Birthday, s *discordgo.Session, g *discordgo.Guild) (*domain.Birthday, error) {

	birthday.Server = domain.DiscordServer{Guild: g, ServerID: g.ID}
	birthday.GuildUser, _ = s.User(birthday.GuildUserID)

	// event.Creator = creator

	return birthday, nil
}

func getBirthdayFromStmt(stmt *sqlite3.Stmt, d *BirthdayDao, g *discordgo.Guild) (*domain.Birthday, error) {
	hasRow, err := stmt.Step()
	if err != nil {
		return nil, err
	}

	if !hasRow {
		return nil, nil
	}
	birthday, err := mapRowToBirthday(stmt, d.Session, g)
	if err != nil {
		return nil, err
	}
	return birthday, err
}

func processStmtForBirthdayArray(stmt *sqlite3.Stmt, d *BirthdayDao, g *discordgo.Guild) ([]*domain.Birthday, error) {
	birthdays := make([]*domain.Birthday, 0)
	for {
		birthday, err := getBirthdayFromStmt(stmt, d, g)
		if err != nil {
			return nil, err
		}
		if birthday != nil {
			birthdays = append(birthdays, birthday)
		} else if len(birthdays) > 0 {
			break
		} else {
			return nil, nil // errors.New("Couldn't find event")
		}
	}
	return birthdays, nil
}

func processStmtForBirthday(stmt *sqlite3.Stmt, d *BirthdayDao, g *discordgo.Guild) (*domain.Birthday, error) {
	birthday, err := getBirthdayFromStmt(stmt, d, g)
	if err != nil {
		stmt.Close()
		return nil, err
	}
	return birthday, err
}
