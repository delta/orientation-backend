package config

// Initializes config package
func InitConfig() {
	initLogger()
	initDB()
	initRDB()

	// clear redis
	RDB.FlushAll()
}
