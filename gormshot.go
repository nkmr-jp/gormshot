package gormshot

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type gormShot struct {
	db          *gorm.DB
	snapshotDir string
	updateFlag  bool // force update flag that snapshot.
}

func New(db *gorm.DB) *gormShot {
	return &gormShot{
		db: db,
	}
}

func (s *gormShot) getSnapshotDir() string {
	if s.snapshotDir == "" {
		return "./.snapshot"
	}
	return s.snapshotDir
}

func (s *gormShot) getSnapshotPath(t *testing.T) string {
	return s.getSnapshotDir() + "/" + strings.Replace(t.Name(), "/", "__", 1) + ".jsonl"
}

func (s *gormShot) SetSnapshotDir(option string) *gormShot {
	s.snapshotDir = option
	return s
}

func (s *gormShot) SetUpdateFlag(option bool) *gormShot {
	s.updateFlag = option
	return s
}

// Save snapshot file save as a JSON Lines text file format.(See: https://jsonlines.org/ )
func (s *gormShot) Save(t *testing.T, model interface{}, selectFields interface{}, order string) (flg bool) {
	t.Helper()
	// gorm query
	rows, err := s.db.Model(model).Find(selectFields).Order(order).Rows()
	if err != nil {
		t.Error(err)
		return false
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			t.Error(err)
			flg = false
		}
	}(rows)

	// make dir and file
	if err := os.MkdirAll(s.getSnapshotDir(), os.ModePerm); err != nil {
		t.Error(err)
		return false
	}

	f, _ := os.Create(s.getSnapshotPath(t))
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Error(err)
			flg = false
		}
	}(f)

	// write each rows as json format.
	w := bufio.NewWriter(f)
	defer func(w *bufio.Writer) {
		if err := w.Flush(); err != nil {
			t.Error(err)
			flg = false
		}
	}(w)
	for rows.Next() {
		if err := s.db.Model(model).ScanRows(rows, selectFields); err != nil {
			t.Error(err)
			return false
		}
		str, err := json.Marshal(selectFields)
		if err != nil {
			t.Error(err)
			return false
		}
		if _, err := w.WriteString(string(str) + "\n"); err != nil {
			t.Error(err)
			return false
		}
	}
	return
}

// Assert asserts that snapshot and db data are equal.
func (s *gormShot) Assert(t *testing.T, model interface{}, selectFields interface{}, order string) (flg bool) {
	t.Helper()
	// Open snapshot file. if no file do save.
	f, err := os.Open(s.getSnapshotPath(t))
	if errors.Is(err, os.ErrNotExist) || s.updateFlag {
		return s.Save(t, model, selectFields, order)
	} else if err != nil {
		t.Error(err)
		return false
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			t.Error(err)
			flg = false
		}
	}(f)

	// When file is exists, assert that snapshot and db data are equal.
	tx := s.db.Model(model).Find(selectFields).Order(order)
	rows, err := tx.Rows()
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			t.Error(err)
			flg = false
		}
	}(rows)
	if err != nil {
		t.Error(err)
		return false
	}

	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		expected := scanner.Text()

		rows.Next()
		if err := s.db.Model(model).ScanRows(rows, selectFields); err != nil {
			t.Error(err)
			return false
		}
		str, err := json.Marshal(selectFields)
		if err != nil {
			t.Error(err)
			return false
		}
		actual := string(str)
		assert.JSONEqf(t, expected, actual, "Diff detected. %v:%v", s.getSnapshotPath(t), line)
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
		return false
	}
	// Assert that snapshot line count and db data count are equal.
	var count int64
	tx.Count(&count)
	assert.Equalf(t, line, int(count), "Data count is not match. %v:%v", s.getSnapshotPath(t), line)

	return
}
