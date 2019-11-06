package main

import (
	"PhotoCollocator/photo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
)

var (
	path      string
	dbName    string
	threshold float64
)

func main() {
	app := cli.NewApp()
	app.Name = "PhotoCollector"
	app.Usage = "TODO"
	app.Commands = []cli.Command{
		{
			Name:   "scan",
			Usage:  "TODO",
			Action: scan,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "path",
					Value:       ".",
					Destination: &path,
				},
				cli.StringFlag{
					Name:        "db-name",
					Value:       "./photo.db",
					Destination: &dbName,
				},
			},
		},
		{
			Name:   "analize",
			Usage:  "TODO",
			Action: analize,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "db-name",
					Value:       "./photo.db",
					Destination: &dbName,
				},
			},
		},
		{
			Name:   "compare",
			Usage:  "TODO",
			Action: compare,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "db-name",
					Value:       "./photo.db",
					Destination: &dbName,
				},
				cli.Float64Flag{
					Name:        "threshold",
					Value:       0.8,
					Destination: &threshold,
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func scan(c *cli.Context) {
	storage, err := photo.GetDB(dbName)
	if err != nil {
		logrus.WithError(err).Fatal("Database init error")
	}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		img := photo.Image{
			FileName: info.Name(),
			Ext:      filepath.Ext(path),
			FilePath: path,
			Size:     int(info.Size()),
		}
		err = storage.InsertImage(&img)
		if err != nil {
			return errors.Wrap(err, "insert image error")
		}
		return nil
	})
}

func analize(c *cli.Context) {
	storage, err := photo.GetDB(dbName)
	if err != nil {
		logrus.WithError(err).Fatal("Database init error")
	}
	err = storage.Walk(true, func(img *photo.Image) error {
		if err := img.UpdateModTime(); err != nil {
			logrus.WithError(err).Warn("Update mod time error")
		}
		if err := img.UpdateExif(); err != nil {
			logrus.WithError(err).Warn("Update exif error")
		}
		if err := img.UpdateHash(); err != nil {
			logrus.WithError(err).Warn("Update hash error")
		}
		storage.UpdateImage(img, []string{"mod_time", "crt_time", "crt_time_exif", "hash"})
		return nil
	})
	if err != nil {
		logrus.WithError(err).Error("Image analize error")
	}
}
func compare(c *cli.Context) {
	storage, err := photo.GetDB(dbName)
	if err != nil {
		logrus.WithError(err).Fatal("Database init error")
	}
	images, err := storage.GetImages(true)
	if err != nil {
		logrus.WithError(err).Fatal("Cannot get images from db")
	}
	dirsDiff, err := photo.CompareDir(images, float32(threshold))
	if err != nil {
		logrus.WithError(err).Fatal("CompareDir error")
	}

	for _, dirDiff := range dirsDiff {
		logrus.Infof("%s", dirDiff)
		if dirDiff.Diff == 1 {
			logrus.Info("deactivate")
			dirDiff.DeactivateIntersection(storage)
		}
	}
}
