package main

import (
	"path/filepath"
	"os"
	"fmt"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"flag"
	"strconv"
	"github.com/djherbis/times"
	"github.com/adassow/goexiftool"
	"crypto/md5"
	"io"
	"encoding/hex"
	"time"
)

var db *sql.DB

func initDB(){
	db, _ = sql.Open("sqlite3", "./foo.db")
	stmt := `CREATE TABLE IF NOT EXISTS images
                 (id INTEGER PRIMARY KEY,
                 file_name TEXT,
                 ext TEXT,
                 crt_time TEXT,
                 mod_time TEXT,
                 exif BLOB,
                 crt_time_exif TEXT,
                 size INTEGER,
                 hash TEXT,
                 file_path TEXT,
                 dest TEXT)`
	db.Exec(stmt)
}
func visit(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		return nil
	}
	stmt, _ := db.Prepare(`INSERT INTO images ( 
	file_name, ext, file_path, size) VALUES
	(?,?,?,?)`)
	stmt.Exec(f.Name(), filepath.Ext(path), path, strconv.FormatInt(f.Size(),10))
	fmt.Printf("Visited: %s\n", path)
	return nil
}
func index(id int, path string) {
	fmt.Printf("Visited index: %s\n", path)
	stat, err := os.Stat(path)
	if stat.IsDir() {
		return
	}
	info, err := times.Stat(path)
	if err != nil {
		fmt.Printf("Stat error: %s\n", path)
		// TODO: handle errors (e.g. file not found)
	}
	crtTime := time.Time{}
	if info.HasBirthTime(){
		crtTime = info.BirthTime()
	} else {
		fmt.Printf("crtTime not found for file %s\n", path)
	}

	exifCreate := time.Time{}
	m, err := goexiftool.NewMediaFile(path)
	if err != nil {
		fmt.Printf("exif not found %s\n", path)
	} else {
		exifCreate, err = m.GetDate()
	}
	if err != nil {
		fmt.Printf("get exif date error %s\n", path)
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Print(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Print(err)
	}

	stmt, _ := db.Prepare(`UPDATE images SET 
	mod_time=?, crt_time=?, crt_time_exif=?, exif=?, size=?, hash=? WHERE id =?`)
	_, err = stmt.Exec(info.ModTime(), crtTime, exifCreate, m.String(), stat.Size(), hex.EncodeToString(h.Sum(nil)), id)
	if err != nil {
		fmt.Printf("Update: %s: %v\n", path, err)
	}
}

type row struct {
	id int
	path string
}
func dbWalk(handle func(int, string)) {
	rows, _ := db.Query("SELECT id, file_path FROM images")

	var bleh []row
	for rows.Next() {
		var path string
		var id int
		_ = rows.Scan(&id, &path)
		bleh = append(bleh, row{id:id, path:path})

	}
	for _, ble := range bleh{
		handle(ble.id, ble.path)
	}
}

func main() {
	initDB()
	root := flag.String("p", ".", "dir path")
	i := flag.Bool("i", false, "index")
	s := flag.Bool("s", false, "status")
	flag.Parse()
	switch {
	case *i:
		filepath.Walk(*root, visit)
	case *s:
		dbWalk(index)
	}
}
func init() {

}