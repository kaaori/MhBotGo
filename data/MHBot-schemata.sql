CREATE TABLE IF NOT EXISTS Servers (
    ServerID        VARCHAR(20) Primary Key,
    JoinTimeUnix    INTEGER
);

CREATE TABLE IF NOT EXISTS Stats (
    StatsID             INTEGER Primary Key,
    ServerID            INTEGER,
    CommandsUsed        INTEGER,
    EventsRun           INTEGER,
    LastEventTimeUnix   INTEGER,
    BirthdaysRegistered INTEGER,

    FOREIGN KEY(ServerID) REFERENCES Servers(ServerID)
);

CREATE TABLE IF NOT EXISTS Events (
    EventID                     INTEGER Primary Key,
    ServerID                    VARCHAR(20),
    CreatorID                   VARCHAR(20),
    EventLocation               VARCHAR(256),
    HostName                    VARCHAR(256),
    CreationTimestamp           INTEGER,
    StartTimestamp              INTEGER,
    LastAnnouncementTimestamp   INTEGER,
    DurationMinutes             INTEGER,

    FOREIGN KEY(ServerID) REFERENCES Servers(ServerID)
);

        CREATE TABLE IF NOT EXISTS PingedRoles(
            RoleID INTEGER Primary Key,
            EventID INTEGER,
            
            FOREIGN KEY(EventID) REFERENCES Events(EventID)
        ); 

CREATE TABLE IF NOT EXISTS Birthdays(
    BirthdayID INTEGER Primary Key,
    ServerID INTEGER,
    UserID VARCHAR(20),
    BirthMonthNum INTEGER,
    BirthDayNum INTEGER,

    FOREIGN KEY(ServerID) REFERENCES Servers(ServerID)
);