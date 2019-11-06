package photo

import (
	"PhotoCollocator/src/github.com/djherbis/times"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oleiade/reflections"
	"github.com/pkg/errors"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

type Image struct {
	Id          int
	FileName    string
	Ext         string
	crtTime     time.Time
	modTime     time.Time
	exif        []byte
	crtTimeExif time.Time
	Size        int
	hash        string
	FilePath    string
	dest        sql.NullString
	active      sql.NullBool
}

func (i *Image)UpdateModTime() error{
	stat, err := os.Stat(i.FilePath)
	if stat.IsDir() {
		return errors.New("It's not a file")
	}

	info, err := times.Stat(i.FilePath)
	if err != nil {
		return errors.Wrapf(err, "Stat error: %s", i.FilePath)
	}
	if info.HasBirthTime() {
		i.crtTime = info.BirthTime()
	} else {
		if info.HasChangeTime() {
			i.crtTime = info.ChangeTime()
		} else {
			fmt.Printf("crtTime not found for file %s\n", i.FilePath)
		}
	}
	i.modTime = info.ModTime()
	return nil
}

func (i *Image)UpdateExif() error{
	f, err := os.Open(i.FilePath)
	if err != nil {
		fmt.Print(err)
	}
	defer f.Close()
	exif.RegisterParsers(mknote.All...)

	if x, err := exif.Decode(f); err != nil {
		return errors.Wrap(err, "exif decode error")
	} else {
		if i.crtTimeExif, err = x.DateTime(); err != nil {
			return errors.Wrap(err, "exif data error")
		}
	}
	return nil
}

func (i *Image)UpdateHash() error{
	f, err := os.Open(i.FilePath)
	if err != nil {
		fmt.Print(err)
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return errors.Wrap(err,"io copy error")
	}
	//Get the 16 bytes hash
	hashInBytes := h.Sum(nil)[:16]

	//Convert the bytes to a string
	i.hash = hex.EncodeToString(hashInBytes)
	return nil
}

type storage struct {
	db *sql.DB
}

func GetDB(dataSourceName string) (*storage, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &storage{db}, nil
}
func (s storage) Init() error {
	stmt := `CREATE TABLE IF NOT EXISTS images
                 (Id INTEGER PRIMARY KEY,
                 file_name TEXT,
                 Ext TEXT,
                 crt_time DATETIME,
                 mod_time DATETIME,
                 exif BLOB,
                 crt_time_exif DATETIME,
                 Size INTEGER,
                 hash TEXT,
                 file_path TEXT,
                 dest TEXT,
                 active BOOLEAN)`
	_, err := s.db.Exec(stmt)
	return err
}

func (s storage) GetImages(active bool) ([]*Image, error) {
	rows, _ := s.db.Query(`SELECT id,file_name,ext,crt_time,mod_time,exif,crt_time_exif,size,hash,file_path,dest,active 
									FROM images 
									WHERE active is NULL or active == ?`, active)
	var images []*Image
	for rows.Next() {
		img := Image{}
		err := rows.Scan(
			&img.Id,
			&img.FileName,
			&img.Ext,
			&img.crtTime,
			&img.modTime,
			&img.exif,
			&img.crtTimeExif,
			&img.Size,
			&img.hash,
			&img.FilePath,
			&img.dest,
			&img.active)
		if err != nil {
			return nil, errors.Wrap(err,"Row scan error")
		} else {
			images = append(images, &img)
		}
	}
	return images, nil
}
func (s storage) Walk(active bool, handle func(*Image) error) error {
	images, err:= s.GetImages(active)
	if err != nil{
		return errors.Wrap(err, "GetImage error")
	}
	for _, img := range images {
		err := handle(img)
		if err != nil {
			logrus.WithError(err).Error("Image handle error")
		}
	}
	return nil
}
func (s storage) DeactivateImages(images []*Image) error {
	for _, img := range images {
		err := s.DeactivateImage(img)
		if err != nil {
			logrus.WithError(err).Error("Deactivate image  error")
		}
	}
	return nil
}

func (s storage) DeactivateImage(img *Image) error {
	stmt, _ := s.db.Prepare(`UPDATE images SET active=false WHERE Id = ?`)
	_, err := stmt.Exec(img.Id)
	return err
}

func (s storage) UpdateImage(img *Image, fields []string) error {
	var fieldsSet []string
	var values []interface{}
	for _, field := range fields{
		value, err := reflections.GetField(s, field)
		if err != nil {
			return errors.Wrap(err,"getField Error")
		}
		fieldsSet = append(fieldsSet, fmt.Sprintf("%s=?", field))
		values = append(values, value)
	}
	stmt, _ := s.db.Prepare(fmt.Sprintf(`UPDATE %s SET %s WHERE Id = ?`, strings.Join(fields,","),strings.Join(fieldsSet,",")))
	values = append(values, img.Id)
	_, err := stmt.Exec(values...)
	return err
}

func (s storage) InsertImage(img *Image) error {
	stmt, _ := s.db.Prepare(`INSERT INTO images (file_name,ext,crt_time,mod_time,exif,crt_time_exif,size,hash,file_path,dest,active) VALUES
	(?,?,?,?,?,?,?,?,?,?,?)`)
	_, err := stmt.Exec(&img.FileName,
		&img.Ext,
		&img.crtTime,
		&img.modTime,
		&img.exif,
		&img.crtTimeExif,
		&img.Size,
		&img.hash,
		&img.FilePath,
		&img.dest,
		&img.active)
	return err
}

func DeactivateDir(db *sql.DB, dir string) error {
	logrus.Infof("deactivate %s", dir)
	stmt, _ := db.Prepare(`UPDATE images SET active=false WHERE file_path LIKE ?`)
	_, err := stmt.Exec(fmt.Sprintf("%%%s%%", dir))
	return err
}
