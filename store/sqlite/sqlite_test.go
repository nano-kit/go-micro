package sqlite

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kr/pretty"
	"github.com/micro/go-micro/v2/store"
)

func TestSQL(t *testing.T) {
	sqlStore := NewStore(
		store.Database("testsql"),
	)
	defer cleanup("testsql", sqlStore)

	if err := sqlStore.Init(); err != nil {
		t.Fatal(err)
	}

	keys, err := sqlStore.List()
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%# v\n", pretty.Formatter(keys))
	}

	err = sqlStore.Write(
		&store.Record{
			Key:   "test",
			Value: []byte("foo"),
		},
	)
	if err != nil {
		t.Error(err)
	}
	err = sqlStore.Write(
		&store.Record{
			Key:   "bar",
			Value: []byte("baz"),
		},
	)
	if err != nil {
		t.Error(err)
	}
	err = sqlStore.Write(
		&store.Record{
			Key:   "qux",
			Value: []byte("aasad"),
		},
	)
	if err != nil {
		t.Error(err)
	}
	err = sqlStore.Delete("qux")
	if err != nil {
		t.Error(err)
	}

	err = sqlStore.Write(&store.Record{
		Key:      "test",
		Value:    []byte("bar"),
		Metadata: map[string]interface{}{"x": 1, "y": 2},
		Expiry:   time.Second * 10,
	})
	if err != nil {
		t.Error(err)
	}

	records, err := sqlStore.Read("test")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%# v\n", pretty.Formatter(records))
	if string(records[0].Value) != "bar" {
		t.Error("Expected bar, got ", string(records[0].Value))
	}

	time.Sleep(11 * time.Second)
	_, err = sqlStore.Read("test")
	switch err {
	case nil:
		t.Error("Key test should have expired")
	default:
		t.Error(err)
	case store.ErrNotFound:
		break
	}
	sqlStore.Delete("bar")
	sqlStore.Write(&store.Record{Key: "aaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	sqlStore.Write(&store.Record{Key: "aaaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	sqlStore.Write(&store.Record{Key: "aaaaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	results, err := sqlStore.Read("a", store.ReadPrefix())
	if err != nil {
		t.Error(err)
	}
	if len(results) != 3 {
		t.Fatal("Results should have returned 3 records")
	}
	time.Sleep(6 * time.Second)
	results, err = sqlStore.Read("a", store.ReadPrefix())
	if err != nil {
		t.Error(err)
	}
	if len(results) != 0 {
		t.Fatal("Results should have returned 0 records")
	}
}

func cleanup(db string, s store.Store) {
	s.Close()
	dir := filepath.Join(DefaultDir, db+".db")
	os.RemoveAll(dir)
}

func TestSqlStoreReInit(t *testing.T) {
	s := NewStore(store.Table("aaa"))
	defer cleanup(DefaultDatabase, s)

	s.Init(store.Table("bbb"))
	if s.Options().Table != "bbb" {
		t.Error("Init didn't reinitialise the store")
	}
	ss := s.(*sqlStore)
	if len(ss.databases) != 1 {
		t.Error("Init ditn't clear last db handle")
	}
	err := s.Write(&store.Record{Key: "foo"}, store.WriteTo(DefaultDatabase, "ccc"))
	if err != nil {
		t.Error(err)
	}
	if len(ss.databases) != 2 {
		t.Error("new table ccc didn't mark")
	}
	bbb := ss.databases[DefaultDatabase+":"+"bbb"]
	ccc := ss.databases[DefaultDatabase+":"+"ccc"]
	if bbb == nil || ccc == nil || bbb != ccc {
		t.Error("two table can not reuse one db handle")
	}
	err = s.Write(&store.Record{Key: "foo"}, store.WriteTo("testtest", "ccc"))
	if err != nil {
		t.Error(err)
	}
	defer cleanup("testtest", s)
	if len(ss.databases) != 3 {
		t.Error("new database testtest table ccc didn't mark")
	}
	ccc1 := ss.databases["testtest"+":"+"ccc"]
	if ccc1 == nil || ccc1 == bbb || ccc1 == ccc {
		t.Error("two database reuse one db handle")
	}
}

func TestSqlStoreList(t *testing.T) {
	sqlStore := NewStore(
		store.Database("testlist"),
	)
	defer cleanup("testlist", sqlStore)

	if err := sqlStore.Init(); err != nil {
		t.Fatal(err)
	}

	sqlStore.Write(&store.Record{Key: "foo", Value: []byte("bar")})
	sqlStore.Write(&store.Record{Key: "aaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	sqlStore.Write(&store.Record{Key: "aaaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	sqlStore.Write(&store.Record{Key: "aaaaa", Value: []byte("bbb"), Expiry: 5 * time.Second})
	results, err := sqlStore.List(store.ListPrefix("a"))
	if err != nil {
		t.Error(err)
	}
	if len(results) != 3 {
		t.Fatal("Results should have returned 3 records")
	}
	results, err = sqlStore.List()
	if err != nil {
		t.Error(err)
	}
	if len(results) != 4 {
		t.Fatal("Results should have returned 4 records")
	}
	time.Sleep(6 * time.Second)
	results, err = sqlStore.List(store.ListPrefix("a"))
	if err != nil {
		t.Error(err)
	}
	if len(results) != 0 {
		t.Fatal("Results should have returned 0 records")
	}
}
