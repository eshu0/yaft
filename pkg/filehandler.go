package yaft

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	sli "github.com/eshu0/simplelogger/pkg/interfaces"
	"github.com/eshu0/yaft/pkg/datastore"
	"github.com/eshu0/yaft/pkg/models"
)

func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func CompareDirectory(fds *datastore.Storage, Logger sli.ISimpleLogger, die *bool, y2a *bool, reader *bufio.Reader) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			Logger.LogErrorE("CompareDirectory", err)
			return nil
		}

		//fd.FilePath = path
		abs, err := filepath.Abs(path)
		if err != nil {
			Logger.LogErrorE("CompareDirectory - Abs", err)
			return nil
		}

		if !info.IsDir() {
			hr := models.HashRelationship{}

			if hr.GenHashData(Logger, abs, info.IsDir()) {
				Logger.LogInfof("CompareDirectory", "This file %s has hashdata %s \n", abs, hr.Hash.Data)

				hr.Path = abs
				res := fds.FindHashData(hr.Hash.Data)
				for _, v := range res {
					hrs := fds.GetHashRelationshipByHash(v.ID)
					for _, hr := range hrs {
						Logger.LogInfo("CompareDirectory", hr.Path)
					}
					if die != nil && *die && len(hrs) > 0 {
						if y2a != nil && *y2a {
							Logger.LogInfof("CompareDirectory", "(y2a) Deleting file %s: \n", path)

							err := os.Remove(path)
							if err != nil {
								Logger.LogErrorE("CompareDirectory - deleting file", err)
								return nil
							} else {
								Logger.LogInfof("CompareDirectory", "deleted file %s\n", path)
							}
						} else {
							fmt.Printf("Delete file %s: \n", path)
							Logger.LogInfof("CompareDirectory", "Delete file %s: \n", path)
							text, _ := reader.ReadString('\n')
							text = strings.Replace(text, "\n", "", -1)
							if strings.Contains(text, "yes") || strings.Contains(text, "y") {
								Logger.LogInfof("CompareDirectory", "Deleting file %s: \n", path)
								err := os.Remove(path)
								if err != nil {
									Logger.LogErrorE("CompareDirectory - deleting file", err)
									return nil
								} else {
									Logger.LogInfof("CompareDirectory", "deleted file %s\n", path)
								}
							} else {
								Logger.LogInfof("CompareDirectory", "Not deleted file %s\n", path)
							}
						}

					}
				}
			}
		}

		return nil
	}
}

func WalkDir(fds *datastore.Storage, Logger sli.ISimpleLogger, persist bool) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {

		if err != nil {
			Logger.LogErrorE("WalkDir", err)
			return nil
		}
		fexts := filepath.Ext(path)

		if strings.ToLower(fexts) != ".yaft" {

			//fd.FilePath = path
			abs, err := filepath.Abs(path)
			if err != nil {
				Logger.LogErrorE("WalkDir - Abs", err)
				return nil
			}

			fwn := FilenameWithoutExtension(abs)
			filename := filepath.Base(abs)
			Logger.LogInfof("WalkDir", "Filename: %s", filename)

			if filename[0] != '.' {
				fwn += ".yaft"
				hd := &models.HashData{}
				ok, _ := hd.CheckFileExists(fwn)
				if ok {
					data, ok := hd.LoadHashData(fwn, Logger)
					if ok {
						hd = data
					}
				} else {
					if !info.IsDir() {
						hr := models.HashRelationship{}

						if hr.GenHashData(Logger, abs, info.IsDir()) {
							Logger.LogInfof("WalkDir", "%s %s\n", abs, hr.Hash.Data)

							hr.Path = abs
							fds.AddHashRelationship(&hr)
							if persist {
								hr.Hash.Save(fwn, Logger)
							}
						}
					}
				}
			} else {
				Logger.LogInfof("WalkDir", "Hidden file %s \n", fwn)
			}

		}

		return nil
		/*
			} else {
				fmt.Printf("%s is directory\n", path)
				Logger.LogInfo("Visit", fmt.Sprintf("%s is directory", path))
				return nil
			}
		*/
	}
}
