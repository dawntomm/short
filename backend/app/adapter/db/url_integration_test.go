// +build integration

package db_test

import (
	"database/sql"
	"fmt"
	"path"
	"short/app/adapter/db"
	"short/app/adapter/db/table"
	"short/app/entity"
	"short/dep"
	"strconv"
	"testing"
	"time"

	"github.com/byliuyang/app/fw"
	"github.com/byliuyang/app/mdtest"
)

var dbConnector fw.DBConnector
var dbMigrationTool fw.DBMigrationTool

var dbConfig fw.DBConfig
var dbMigrationRoot string

var insertRowSQL = fmt.Sprintf(`
INSERT INTO %s (%s, %s, %s, %s, %s)
VALUES ($1, $2, $3, $4, $5)`,
	table.URL.TableName,
	table.URL.ColumnAlias,
	table.URL.ColumnOriginalURL,
	table.URL.ColumnCreatedAt,
	table.URL.ColumnExpireAt,
	table.URL.ColumnUpdatedAt,
)

type tableRow struct {
	alias     string
	longLink  string
	createdAt time.Time
	expireAt  time.Time
	updatedAt time.Time
}

func TestURLSql_IsAliasExist(t *testing.T) {
	testCases := []struct {
		name       string
		tableRows  []tableRow
		alias      string
		expIsExist bool
	}{
		{
			name:       "alias doesn't exist",
			alias:      "gg",
			tableRows:  []tableRow{},
			expIsExist: false,
		},
		{
			name:  "alias found",
			alias: "gg",
			tableRows: []tableRow{
				{alias: "gg"},
			},
			expIsExist: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mdtest.AccessTestDB(
				dbConnector,
				dbMigrationTool,
				dbMigrationRoot,
				dbConfig,
				func(sqlDB *sql.DB) {
					insertTableRows(t, sqlDB, testCase.tableRows)

					urlRepo := db.NewURLSql(sqlDB)
					gotIsExist, err := urlRepo.IsAliasExist(testCase.alias)
					mdtest.Equal(t, nil, err)
					mdtest.Equal(t, testCase.expIsExist, gotIsExist)
				})
		})
	}
}

func TestURLSql_GetByAlias(t *testing.T) {
	twoYearsAgo := mustParseTime(t, "2017-05-01T08:02:16Z")
	now := mustParseTime(t, "2019-05-01T08:02:16Z")

	testCases := []struct {
		name        string
		tableRows   []tableRow
		alias       string
		hasErr      bool
		expectedURL entity.URL
	}{
		{
			name:      "alias not found",
			tableRows: []tableRow{},
			alias:     "220uFicCJj",
			hasErr:    true,
		},
		{
			name: "found url",
			tableRows: []tableRow{
				{
					alias:     "220uFicCJj",
					longLink:  "http://www.google.com",
					createdAt: twoYearsAgo,
					expireAt:  now,
				},
				{
					alias:     "yDOBcj5HIPbUAsw",
					longLink:  "http://www.facebook.com",
					createdAt: twoYearsAgo,
					expireAt:  now,
				},
			},
			alias:  "220uFicCJj",
			hasErr: false,
			expectedURL: entity.URL{
				Alias:       "220uFicCJj",
				OriginalURL: "http://www.google.com",
				CreatedAt:   &twoYearsAgo,
				ExpireAt:    &now,
				UpdatedAt:   nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mdtest.AccessTestDB(
				dbConnector,
				dbMigrationTool,
				dbMigrationRoot,
				dbConfig,
				func(sqlDB *sql.DB) {
					insertTableRows(t, sqlDB, testCase.tableRows)

					urlRepo := db.NewURLSql(sqlDB)
					url, err := urlRepo.GetByAlias(testCase.alias)

					if testCase.hasErr {
						mdtest.NotEqual(t, nil, err)
						return
					}
					mdtest.Equal(t, nil, err)
					mdtest.Equal(t, testCase.expectedURL, url)
				},
			)
		})
	}
}

func TestURLSql_Create(t *testing.T) {
	now := mustParseTime(t, "2019-05-01T08:02:16Z")

	testCases := []struct {
		name      string
		tableRows []tableRow
		url       entity.URL
		hasErr    bool
	}{
		{
			name: "alias exists",
			tableRows: []tableRow{
				{
					alias:    "220uFicCJj",
					longLink: "http://www.facebook.com",
					expireAt: now,
				},
			},
			url: entity.URL{
				Alias:       "220uFicCJj",
				OriginalURL: "http://www.google.com",
				ExpireAt:    &now,
			},
			hasErr: true,
		},
		{
			name: "successfully create url",
			tableRows: []tableRow{
				{
					alias:    "abc",
					longLink: "http://www.google.com",
					expireAt: now,
				},
			},
			url: entity.URL{
				Alias:       "220uFicCJj",
				OriginalURL: "http://www.google.com",
				ExpireAt:    &now,
			},
			hasErr: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mdtest.AccessTestDB(
				dbConnector,
				dbMigrationTool,
				dbMigrationRoot,
				dbConfig,
				func(sqlDB *sql.DB) {
					urlRepo := db.NewURLSql(sqlDB)
					err := urlRepo.Create(testCase.url)

					if testCase.hasErr {
						mdtest.NotEqual(t, nil, err)
						return
					}
					mdtest.Equal(t, nil, err)
				},
			)
		})
	}
}

func insertTableRows(t *testing.T, sqlDB *sql.DB, tableRows []tableRow) {
	for _, tableRow := range tableRows {
		_, err := sqlDB.Exec(
			insertRowSQL,
			tableRow.alias,
			tableRow.longLink,
			tableRow.createdAt,
			tableRow.expireAt,
			tableRow.updatedAt,
		)
		mdtest.Equal(t, nil, err)
	}
}

func mustParseTime(t *testing.T, timeString string) time.Time {
	parsedTime, err := time.Parse(time.RFC3339, timeString)
	mdtest.Equal(t, nil, err)
	return parsedTime
}

func TestMain(m *testing.M) {
	env := dep.InjectEnvironment()
	env.AutoLoadDotEnvFile()

	host := env.GetEnv("DB_HOST", "")
	portStr := env.GetEnv("DB_PORT", "")
	port := mustInt(portStr)
	user := env.GetEnv("DB_USER", "")
	password := env.GetEnv("DB_PASSWORD", "")
	dbName := env.GetEnv("DB_NAME", "")

	dbConfig = fw.DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
	}

	dbMigrationRoot = path.Join(env.GetEnv("MIGRATION_ROOT", ""))

	dbConnector = dep.InjectDBConnector()
	dbMigrationTool = dep.InjectDBMigrationTool()

	m.Run()
}

func mustInt(numStr string) int {
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num
}
