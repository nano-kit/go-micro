// Package sqlite implements the sqlite store
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/store"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

// DefaultDatabase is the namespace that the sql store
// will use if no namespace is provided.
var (
	DefaultDatabase = "micro"
	DefaultTable    = "micro"
	// DefaultDir is the default directory for sqlite files
	DefaultDir = filepath.Join(homeDir(), ".microsqlstore")
)

var (
	re = regexp.MustCompile("[^a-zA-Z0-9]+")

	statements = map[string]string{
		"list":       "SELECT key, value, metadata, expiry FROM %s;",
		"read":       "SELECT key, value, metadata, expiry FROM %s WHERE key = $1;",
		"readMany":   "SELECT key, value, metadata, expiry FROM %s WHERE key LIKE $1;",
		"readOffset": "SELECT key, value, metadata, expiry FROM %s WHERE key LIKE $1 ORDER BY key DESC LIMIT $2 OFFSET $3;",
		"write":      "INSERT INTO %s(key, value, metadata, expiry) VALUES ($1, $2, $3, $4) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, metadata = EXCLUDED.metadata, expiry = EXCLUDED.expiry;",
		"delete":     "DELETE FROM %s WHERE key = $1;",
	}
)

func homeDir() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return dir
	}
	return os.TempDir()
}

type sqlStore struct {
	options store.Options
	dir     string

	// the database handle
	sync.RWMutex
	databases map[string]*sql.DB
}

func key(database, table string) string {
	return database + ":" + table
}

// NewStore returns a new micro Store backed by sql
func NewStore(opts ...store.Option) store.Store {
	options := store.Options{
		Database: DefaultDatabase,
		Table:    DefaultTable,
	}

	for _, o := range opts {
		o(&options)
	}

	// new store
	s := new(sqlStore)
	// set the options
	s.options = options
	// mark known databases
	s.databases = make(map[string]*sql.DB)
	// best-effort configure the store
	if err := s.configure(); err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Error("Error configuring store ", err)
		}
	}

	// return store
	return s
}

func (s *sqlStore) Close() error {
	s.Lock()
	defer s.Unlock()
	for k, v := range s.databases {
		v.Close()
		delete(s.databases, k)
	}
	return nil
}

// Delete records with keys
func (s *sqlStore) Delete(key string, opts ...store.DeleteOption) error {
	var options store.DeleteOptions
	for _, o := range opts {
		o(&options)
	}

	// create the db if not exists
	db, err := s.createDB(options.Database, options.Table)
	if err != nil {
		return err
	}

	st, err := s.prepare(db, options.Database, options.Table, "delete")
	if err != nil {
		return err
	}
	defer st.Close()

	result, err := st.Exec(key)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlStore) Init(opts ...store.Option) error {
	for _, o := range opts {
		o(&s.options)
	}
	// reconfigure
	return s.configure()
}

// List all the known records
func (s *sqlStore) List(opts ...store.ListOption) ([]string, error) {
	var options store.ListOptions
	for _, o := range opts {
		o(&options)
	}

	// create the db if not exists
	db, err := s.createDB(options.Database, options.Table)
	if err != nil {
		return nil, err
	}

	st, err := s.prepare(db, options.Database, options.Table, "list")
	if err != nil {
		return nil, err
	}
	defer st.Close()

	rows, err := st.Query()
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var keys []string
	var timehelper sql.NullTime

	for rows.Next() {
		record := &store.Record{}
		metadata := make(Metadata)

		if err := rows.Scan(&record.Key, &record.Value, &metadata, &timehelper); err != nil {
			return keys, err
		}

		// set the metadata
		record.Metadata = toMetadata(&metadata)

		if timehelper.Valid {
			if timehelper.Time.Before(time.Now()) {
				// record has expired
				go s.Delete(record.Key)
			} else {
				record.Expiry = time.Until(timehelper.Time)
				keys = append(keys, record.Key)
			}
		} else {
			keys = append(keys, record.Key)
		}

	}
	rowErr := rows.Close()
	if rowErr != nil {
		// transaction rollback or something
		return keys, rowErr
	}
	if err := rows.Err(); err != nil {
		return keys, err
	}
	return keys, nil
}

func (s *sqlStore) Options() store.Options {
	return s.options
}

// Read a single key
func (s *sqlStore) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	var options store.ReadOptions
	for _, o := range opts {
		o(&options)
	}

	// create the db if not exists
	db, err := s.createDB(options.Database, options.Table)
	if err != nil {
		return nil, err
	}

	if options.Prefix || options.Suffix {
		return s.read(db, key, options)
	}

	var records []*store.Record
	var timehelper sql.NullTime

	st, err := s.prepare(db, options.Database, options.Table, "read")
	if err != nil {
		return nil, err
	}
	defer st.Close()

	row := st.QueryRow(key)
	record := &store.Record{}
	metadata := make(Metadata)

	if err := row.Scan(&record.Key, &record.Value, &metadata, &timehelper); err != nil {
		if err == sql.ErrNoRows {
			return records, store.ErrNotFound
		}
		return records, err
	}

	// set the metadata
	record.Metadata = toMetadata(&metadata)

	if timehelper.Valid {
		if timehelper.Time.Before(time.Now()) {
			// record has expired
			go s.Delete(key)
			return records, store.ErrNotFound
		}
		record.Expiry = time.Until(timehelper.Time)
		records = append(records, record)
	} else {
		records = append(records, record)
	}

	return records, nil
}

func (s *sqlStore) String() string {
	return "sqlite"
}

// Write records
func (s *sqlStore) Write(r *store.Record, opts ...store.WriteOption) error {
	var options store.WriteOptions
	for _, o := range opts {
		o(&options)
	}

	// create the db if not exists
	db, err := s.createDB(options.Database, options.Table)
	if err != nil {
		return err
	}

	st, err := s.prepare(db, options.Database, options.Table, "write")
	if err != nil {
		return err
	}
	defer st.Close()

	metadata := make(Metadata)
	for k, v := range r.Metadata {
		metadata[k] = v
	}

	if r.Expiry != 0 {
		_, err = st.Exec(r.Key, r.Value, metadata, time.Now().Add(r.Expiry))
	} else {
		_, err = st.Exec(r.Key, r.Value, metadata, nil)
	}

	if err != nil {
		return errors.Wrap(err, "Couldn't insert record "+r.Key)
	}

	return nil
}

func (s *sqlStore) createDB(database, table string) (*sql.DB, error) {
	database, table = s.getDB(database, table)

	k := key(database, table)
	s.RLock()
	db, ok := s.databases[k]
	s.RUnlock()
	if ok {
		return db, nil
	}

	s.Lock()
	defer s.Unlock()

	db, err := s.initDB(database, table)
	if err != nil {
		return nil, err
	}

	s.databases[k] = db
	return db, nil
}

func (s *sqlStore) getDB(database, table string) (string, string) {
	if len(database) == 0 {
		if len(s.options.Database) > 0 {
			database = s.options.Database
		} else {
			database = DefaultDatabase
		}
	}

	if len(table) == 0 {
		if len(s.options.Table) > 0 {
			table = s.options.Table
		} else {
			table = DefaultTable
		}
	}

	// store.namespace must only contain letters, numbers and underscores
	database = re.ReplaceAllString(database, "_")
	table = re.ReplaceAllString(table, "_")

	return database, table
}

// initDB is called in an exclusive lock
func (s *sqlStore) initDB(database, table string) (db *sql.DB, err error) {
	// find an existing db handle
	for k, v := range s.databases {
		if strings.HasPrefix(k, database+":") {
			db = v
			break
		}
	}

	var newdb *sql.DB
	if db == nil {
		// create a directory
		dir := DefaultDir
		// create the database handle
		fname := database + ".db"
		// make the dir
		os.MkdirAll(dir, 0700)
		// database path
		dbPath := filepath.Join(dir, fname)
		// sqlite URI filename
		dbURI := "file://" + dbPath

		// create new db handle
		newdb, err = sql.Open("sqlite", dbURI)
		if err != nil {
			return nil, err
		}
		db = newdb
	}

	// close the newly created db upon err
	defer func() {
		if err != nil && newdb != nil {
			newdb.Close()
		}
	}()

	// create table
	_, err = db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s
	(
		key TEXT NOT NULL,
		value BLOB,
		metadata TEXT,
		expiry TIMESTAMP,
		CONSTRAINT %s_pkey PRIMARY KEY (key)
	);`, table, table))
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't create table")
	}

	// create index
	_, err = db.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "%s" ON %s ("key");`, "key_index_"+table, table))
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (s *sqlStore) prepare(db *sql.DB, database, table, query string) (*sql.Stmt, error) {
	st, ok := statements[query]
	if !ok {
		return nil, errors.New("unsupported statement")
	}

	// get DB
	_, table = s.getDB(database, table)

	q := fmt.Sprintf(st, table)
	stmt, err := db.Prepare(q)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}

// Read Many records
func (s *sqlStore) read(db *sql.DB, key string, options store.ReadOptions) ([]*store.Record, error) {
	pattern := "%"
	if options.Prefix {
		pattern = key + pattern
	}
	if options.Suffix {
		pattern = pattern + key
	}

	var rows *sql.Rows
	var err error

	if options.Limit != 0 {
		st, err := s.prepare(db, options.Database, options.Table, "readOffset")
		if err != nil {
			return nil, err
		}
		defer st.Close()

		rows, err = st.Query(pattern, options.Limit, options.Offset)
	} else {
		st, err := s.prepare(db, options.Database, options.Table, "readMany")
		if err != nil {
			return nil, err
		}
		defer st.Close()

		rows, err = st.Query(pattern)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			return []*store.Record{}, nil
		}
		return []*store.Record{}, errors.Wrap(err, "sqlStore.read failed")
	}

	defer rows.Close()

	var records []*store.Record
	var timehelper sql.NullTime

	for rows.Next() {
		record := &store.Record{}
		metadata := make(Metadata)

		if err := rows.Scan(&record.Key, &record.Value, &metadata, &timehelper); err != nil {
			return records, err
		}

		// set the metadata
		record.Metadata = toMetadata(&metadata)

		if timehelper.Valid {
			if timehelper.Time.Before(time.Now()) {
				// record has expired
				go s.Delete(record.Key)
			} else {
				record.Expiry = time.Until(timehelper.Time)
				records = append(records, record)
			}
		} else {
			records = append(records, record)
		}
	}
	rowErr := rows.Close()
	if rowErr != nil {
		// transaction rollback or something
		return records, rowErr
	}
	if err := rows.Err(); err != nil {
		return records, err
	}

	return records, nil
}

func (s *sqlStore) configure() error {
	// clear
	s.Close()

	// get DB
	database, table := s.getDB(s.options.Database, s.options.Table)
	k := key(database, table)

	s.Lock()
	defer s.Unlock()

	// initialise the database
	db, err := s.initDB(database, table)
	if err != nil {
		return err
	}

	s.databases[k] = db
	return nil
}
