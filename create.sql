CREATE TABLE IF NOT EXISTS user (
	uuid TEXT NOT NULL PRIMARY KEY,
	firstname TEXT NOT NULL,
	lastname TEXT NOT NULL,
	email TEXT NOT NULL,
	phone TEXT NOT NULL,
	createdat INTEGER NOT NULL
);