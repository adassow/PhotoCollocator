package photo

import (
	"crypto/md5"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"path"
	"sort"
	"strings"
)

func CompareDir(images []*Image, threshold float32) ([]*dirDiff, error) {
	imageDirMap := getImagesByDir(images)
	equalDir := make(map[string]*dirDiff)
	for dir1, dir1Img := range imageDirMap {
		for dir2, dir2Img := range imageDirMap {
			if dir1 != dir2 {
				dirDiffResult, err := GetDirDiff(dir1Img, dir2Img)
				if err != nil {
					logrus.WithError(err).Errorf("Get dir Diff error")
					continue
				}
				diff := float32(len(dirDiffResult.intersection)) / (float32(len(dir1Img)+len(dir2Img)) / 2)
				if diff > threshold {
					if _, exist := equalDir[dirDiffResult.hash()]; !exist {
						equalDir[dirDiffResult.hash()] = dirDiffResult
					}
				}
			}
		}
	}
	result := make([]*dirDiff, 0, len(equalDir))
	logrus.Infof("%d",len(equalDir))
	for _, diff := range equalDir {
		result = append(result, diff)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Diff < result[j].Diff
	})
	return result, nil
}

type dirDiff struct {
	dirA         *string
	dirB         *string
	aMinusB      []*Image
	bMinusA      []*Image
	intersection []*Image
	Diff         float32
}

func (d *dirDiff) String() string {
	aMinusBstr := make([]string, 0, len(d.aMinusB))
	bMinusAstr := make([]string, 0, len(d.bMinusA))
	for _, img := range d.aMinusB {
		aMinusBstr = append(aMinusBstr, img.FileName)
	}
	for _, img := range d.bMinusA {
		bMinusAstr = append(bMinusAstr, img.FileName)
	}

	return fmt.Sprintf("dir:%s vs dir:%s Diff:%f \n aMinusB:%s\n bMinusA:%s",
		*d.dirA, *d.dirB, d.Diff, strings.Join(aMinusBstr,","), strings.Join(bMinusAstr,","))
}

func (d *dirDiff) DeactivateIntersection(storage *storage) error {
		return storage.DeactivateImages(d.intersection)
}

func GetDirDiff(dirImgA, dirImgB []*Image) (*dirDiff, error) {
	diff := dirDiff{}
	for _, img := range dirImgA {
		dir, _ := path.Split(img.FilePath)
		if diff.dirA == nil {
			diff.dirA = &dir
			continue
		}
		if dir != *diff.dirA {
			return nil, errors.Errorf("dirImgA contain files from two different directories")
		}
	}
	for _, img := range dirImgB {
		dir, _ := path.Split(img.FilePath)
		if diff.dirB == nil {
			diff.dirB = &dir
			continue
		}
		if dir != *diff.dirB {
			return nil, errors.Errorf("dirImgB contain files from two different directories")
		}
	}
	dirImgAMap := imgSliceToMap(dirImgA)
	dirImgBMap := imgSliceToMap(dirImgB)
	diff.intersection = intersection(dirImgAMap, dirImgBMap)
	diff.aMinusB = difference(dirImgAMap, dirImgBMap)
	diff.bMinusA = difference(dirImgBMap, dirImgAMap)
	if len(diff.aMinusB) == 0 && len(diff.bMinusA) == 0 {
		logrus.Info("cyce")
	}
	diff.Diff = float32(len(diff.intersection)) / (float32(len(dirImgA)+len(dirImgB)) / 2)
	return &diff, nil
}

func (d *dirDiff) hash() string {
	dirs := []string{*d.dirA, *d.dirB}
	sort.Strings(dirs)
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s%s", dirs[0], dirs[1]))))
}

func getImagesByDir(images []*Image) map[string][]*Image {
	imageMap := make(map[string][]*Image)
	for _, imageObj := range images {
		dir, _ := path.Split(imageObj.FilePath)
		imageMap[dir] = append(imageMap[dir], imageObj)
	}
	return imageMap
}

func imgSliceToMap(images []*Image) map[string]*Image {
	result := make(map[string]*Image)
	for _, img := range images {
		result[img.hash] = img
	}
	return result
}

func intersection(imgMap1 map[string]*Image, imgMap2 map[string]*Image) []*Image {
	var intersectionSlice []*Image
	for k, img := range imgMap1 {
		if _, ok := imgMap2[k]; ok {
			intersectionSlice = append(intersectionSlice, img)
		}
	}
	return intersectionSlice
}

// imgMap1 - imgMap2
func difference(imgMap1 map[string]*Image, imgMap2 map[string]*Image) []*Image {
	var differenceSlice []*Image
	for k, img := range imgMap1 {
		if _, ok := imgMap2[k]; !ok {
			differenceSlice = append(differenceSlice, img)
		}
	}
	return differenceSlice
}
