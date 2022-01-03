package gormshot_test

import (
	"log"
	"os"
	"testing"

	"github.com/nkmr-jp/gormshot"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

type ProductSnap struct {
	Code  string
	Price uint
}

type User struct {
	gorm.Model
	Name string
	Age  uint
}

type UserSnap struct {
	Name string
	Age  uint
}

func TestMain(m *testing.M) {
	// setup
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Println(err)
	}
	resetTable()
	m.Run()
}

func resetTable() {
	if err := db.Migrator().DropTable(&Product{}, &User{}); err != nil {
		log.Println(err)
	}
	if err := db.AutoMigrate(&Product{}, &User{}); err != nil {
		log.Println(err)
	}
	log.Println("reset table")
}

func resetFile(name string, value []byte) {
	if err := os.WriteFile(name, value, 0644); err != nil {
		log.Println(err)
	}
}

func TestSave(t *testing.T) {
	db.Create(&Product{Code: "D42", Price: 100})
	db.Create(&Product{Code: "D43", Price: 110})

	dir := "./.snapshot_test_save"
	if err := os.RemoveAll(dir); err != nil {
		log.Println(err)
	}
	shot := gormshot.New(db).SetSnapshotDir(dir)
	shot.Save(t, &Product{}, &ProductSnap{}, "code")

	expected := `{"Code":"D42","Price":100}
{"Code":"D43","Price":110}
`
	actual, _ := os.ReadFile(dir + "/TestSave.jsonl")
	assert.Equal(t, expected, string(actual))

	t.Run("when Run", func(t *testing.T) {
		shot.Save(t, &Product{}, &ProductSnap{}, "code")
		actual, _ = os.ReadFile(dir + "/TestSave__when_Run.jsonl")
		assert.Equal(t, expected, string(actual))
	})
}

func TestAssert(t *testing.T) {
	t.Run("value is match", func(t *testing.T) {
		resetTable()
		db.Create(&User{Name: "Alice", Age: 20})
		db.Create(&User{Name: "Bob", Age: 45})
		db.Create(&User{Name: "Carol", Age: 31})
		shot := gormshot.New(db)
		shot.Assert(t, &User{}, &UserSnap{}, "name desc,age")
	})
	t.Run("update snap", func(t *testing.T) {
		resetTable()
		db.Create(&User{Name: "Alice", Age: 21})
		snapFile := "./.snapshot/TestAssert__update_snap.jsonl"
		before, _ := os.ReadFile(snapFile)
		assert.Equal(t, `{"Name":"Alice","Age":20}`+"\n", string(before))

		shot := gormshot.New(db).SetUpdateFlag(true)
		shot.Assert(t, &User{}, &UserSnap{}, "name desc,age")

		after, _ := os.ReadFile(snapFile)
		assert.Equal(t, `{"Name":"Alice","Age":21}`+"\n", string(after))

		resetFile(snapFile, before)
	})
}
