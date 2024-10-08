package mysql

import (
	"os"
	"time"

	"github.com/adisnuhic/go-graphql/config"
	"github.com/adisnuhic/go-graphql/pkg/log"
	"github.com/jinzhu/gorm"

	// Initialize mysql driver
	_ "github.com/go-sql-driver/mysql"
	// Initialize mysql migrate
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"

	// Initialize mysql migrate source file
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// db app data store
var dbStore Store

// Init initialize db
func Init(cfg *config.AppConfig, logger log.ILogger) {
	env := os.Getenv("ENV")
	dbStore = initDB(cfg.DBConnections[env], logger)
	runMigrate(dbStore, logger)
}

// Connection get databse connection
func Connection() Store {
	return dbStore
}

// Close database connection
func Close() error {
	if dbStore != nil {
		return dbStore.DB().Close()
	}
	return nil
}

// initDB init database connection
func initDB(dbConn config.DBConnection, logger log.ILogger) Store {

	if dbConn.DBDialect == "" || dbConn.DBConnection == "" {
		return nil
	}

	// open DB connection
	myDB, err := gorm.Open(dbConn.DBDialect, dbConn.DBConnection)
	if err != nil {
		logger.Errorf("unable to connect to database: %v", err)
	}

	// ping database
	if err := myDB.DB().Ping(); err != nil {
		logger.Errorf("unable to ping database: %v", err)
	}

	// SetMaxIdleConns sets maximum number of connections in the idle connection pool
	maxConn := dbConn.DbMaxIdleConns
	myDB.DB().SetMaxIdleConns(maxConn)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	maxConn = dbConn.DbMaxOpenConns
	myDB.DB().SetMaxOpenConns(maxConn)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	maxConn = dbConn.DbConnMaxLifetime
	duration := time.Minute * time.Duration(maxConn)
	myDB.DB().SetConnMaxLifetime(duration)

	// Enable Logger, show detailed log
	myDB.LogMode(dbConn.DbLogging)

	logger.Print("initialized API database successfully")

	return myDB

}

// executes migrations against database
func runMigrate(store Store, logger log.ILogger) {
	driver, err := mysqlmigrate.WithInstance(store.DB(), &mysqlmigrate.Config{})
	if err != nil {
		logger.Fatalf("unable to migrate: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql", driver)
	if err != nil {
		logger.Fatalf("unable to migrate: %v", err)
	}

	migrateErr := m.Up()
	if migrateErr != nil {
		logger.Printf("unable to migrate: %v", migrateErr)
	}

	if migrateErr == nil {
		logger.Print("migrations executed successfully")
	}

}
