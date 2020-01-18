package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

type ErrorsResponse []ErrorResponse

type UploadResponse struct {
	URL string `json:"url"`
}

type UploadsResponse []UploadResponse

// type Commit struct {
// 	ProjectPath string `bson:"project_path"`
// 	ProjectName string `bson:"project_name"`
// 	projectPath string `bson:"project_path"`
// 	Hash        string `bson:"hash"`
// 	Time        int64  `bson:"time"`
// 	ID          int    `bson:"id"`
// 	// Files       map[string]string `bson:"files"`
// }

// type Commits []Commit

var liveStart bool

const CUI_LOG = ".cui.log"

// const LIVE_CODING_PATH = "/Users/kitamurataku/work/liveCoding"

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func compress(src string, buf io.Writer) error {
	// tar > gzip > buf
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	// walk through every file in the folder
	filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		// must provide real name
		// (see https://golang.org/src/archive/tar/common.go?#L626)
		header.Name = filepath.ToSlash(file)

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})

	// produce tar
	if err := tw.Close(); err != nil {
		return err
	}
	// produce gzip
	if err := zr.Close(); err != nil {
		return err
	}
	//
	return nil
}

func remove(strings []string, search string) []string {
	result := []string{}
	for _, v := range strings {
		if v != search {
			result = append(result, v)
		}
	}
	return result
}

func writeCommandInput(input string, projectPath string, liveStart bool) {
	if liveStart == false {
		return
	}
	file, err := os.OpenFile(projectPath+"/"+CUI_LOG, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(file, "$ "+input)
	file.Close()
}

func writeCommandOut(out string, projectPath string, liveStart bool) {
	fmt.Print(out)
	if liveStart == false {
		return
	}
	file, err := os.OpenFile(projectPath+"/"+CUI_LOG, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(file, out)
	file.Close()
}

func liveCommandUsage(projectPath string) {
	out := "usage: live [start, stop]\n"
	writeCommandOut(out, projectPath, false)
}

func watch(r *git.Repository, projectPath string, couterHTMLPath string) error {
	w, err := r.Worktree()
	if err != nil {
		fmt.Println(err)
		return err
	}

	for {
		if liveStart == false {
			return nil
		}

		time.Sleep(time.Second * 1)

		status, err := w.Status()
		if err != nil {
			fmt.Println(err)
			return err
		}
		// fmt.Println(status.String())

		if len(status) != 0 {
			err = w.AddGlob(".")
			if err != nil {
				fmt.Println(err)
				return err
			}

			commit, err := w.Commit(strconv.FormatInt(time.Now().UnixNano(), 10), &git.CommitOptions{
				Author: &object.Signature{
					When: time.Now(),
				},
			})
			if err != nil {
				fmt.Println(err)
				return err
			}

			w.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName("master"),
			})

			_, err = r.CommitObject(commit)
			if err != nil {
				fmt.Println(err)
				return err
			}

			cIter, err := r.Log(&git.LogOptions{Order: git.LogOrderCommitterTime})
			if err != nil {
				fmt.Println(err)
				return err
			}

			commitID := -1
			err = cIter.ForEach(func(_ *object.Commit) error {
				commitID++
				return nil
			})
			if err != nil {
				fmt.Println(err)
				return err
			}
			// fmt.Println(commitID)

			err, _ = createCounterHTML("ID: "+strconv.Itoa(commitID), couterHTMLPath)
			if err != nil {
				fmt.Println(err)
				return err
			}

			// _, err = r.CommitObject(commit) 以下に貼り付ければライブモード
			// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			// client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			// if err != nil {
			// 	continue
			// }
			// defer client.Disconnect(ctx)

			// commitTime, err := strconv.ParseInt(obj.Message, 10, 64)
			// if err != nil {
			// 	continue
			// }

			// commitCollection := client.Database("liveCoding").Collection("commit")

			// findOptions := options.Find()
			// findOptions.SetSort(bson.D{{"id", -1}}).SetLimit(1)
			// projectFilter := bson.M{"project_path": projectPath}
			// cur, err := commitCollection.Find(ctx, projectFilter, findOptions)
			// if err != nil {
			// 	continue
			// }

			// commitsStruct := Commits{}

			// err = cur.All(ctx, &commitsStruct)
			// if err != nil {
			// 	continue
			// }

			// var commitStruct Commit

			// if len(commitsStruct) == 1 {
			// 	commitStruct = Commit{
			// 		ProjectPath: projectPath,
			// 		ProjectName: filepath.Base(projectPath),
			// 		Hash:        obj.Hash.String(),
			// 		Time:        commitTime,
			// 		ID:          commitsStruct[0].ID + 1,
			// 	}

			// 	err, _ := createCounterHTML("ID: "+strconv.Itoa(commitsStruct[0].ID+1), couterHTMLPath)
			// 	if err != nil {
			// 		return err
			// 	}
			// } else {
			// 	commitStruct = Commit{
			// 		ProjectPath: projectPath,
			// 		ProjectName: filepath.Base(projectPath),
			// 		Hash:        obj.Hash.String(),
			// 		Time:        commitTime,
			// 		ID:          0,
			// 	}

			// 	err, _ := createCounterHTML("ID: 0", couterHTMLPath)
			// 	if err != nil {
			// 		fmt.Println(err)
			// 		return err
			// 	}
			// }

			// _, err = commitCollection.InsertOne(ctx, commitStruct)
			// if err != nil {
			// 	continue
			// }

		}
	}
}

func createCounterHTML(message string, path string) (error, string) {
	// tmpFile, err := ioutil.TempFile("counter.html", "counter.html")
	// if err != nil {
	// 	return err
	// }
	// tmpPath := tmpFile.Name()
	// fmt.Println(tmpPath)

	// fmt.Println(message, path)
	tmpPath := ""

	if path == "" {
		dir, err := ioutil.TempDir("", "")
		if err != nil {
			return err, ""
		}
		tmpPath = filepath.Join(dir, "counter.html")
	} else {
		tmpPath = path
	}

	counterHTML := `<title>LiveCoding</title><style type="text/css">html,body{width:100%;height:100%}html{display:table}body{display:table-cell;text-align:center;vertical-align:middle}#message{font-size:10em}</style><p id="message">` + message + `</p> <script>setTimeout(location.reload(),1000);</script>`

	err := ioutil.WriteFile(tmpPath, []byte(counterHTML), 0666)
	if err != nil {
		return err, ""
	}

	return nil, tmpPath
}

func main() {
	projectPath := ""
	buffer := &bytes.Buffer{}
	writer := buffer
	liveStart = false

	fmt.Println("\x1b[32mWelcome Live Coding Capture! (v0.0.1)\x1b[0m")
	err, couterHTMLPath := createCounterHTML("実況準備中", "")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Please open \"file://" + couterHTMLPath + "\" in your browser.")
	for {
		pwd, err := os.Getwd()
		if err != nil {
			writeCommandOut(err.Error()+"\n", projectPath, liveStart)
			continue
		}

		home, err := os.UserHomeDir()
		if err != nil {
			writeCommandOut(err.Error()+"\n", projectPath, liveStart)
			continue
		}

		var liveStatus string
		if liveStart {
			liveStatus = "recording"
		} else {
			liveStatus = "stopped"
		}

		currentPath := strings.Replace(pwd, home, "~", 1)
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("\x1b[34m%s\x1b[0m \x1b[31m(%s)\x1b[0m %s ", currentPath, liveStatus, "$")
		scanner.Scan()
		line := scanner.Text()

		cmdSplit := strings.Split(line, " ")
		cmdSplit = remove(cmdSplit, "")

		if len(cmdSplit) > 0 {
			firstCommandName := cmdSplit[0]
			if firstCommandName == "cd" {
				writeCommandInput(line, projectPath, liveStart)
				if len(cmdSplit) == 1 {
					home, err := os.UserHomeDir()
					if err != nil {
						writeCommandOut(err.Error()+"\n", projectPath, liveStart)
						continue
					}
					err = os.Chdir(home)
					if err != nil {
						writeCommandOut(err.Error()+"\n", projectPath, liveStart)
						continue
					}
					continue
				} else if len(cmdSplit) == 2 {
					secondCommandValue := cmdSplit[1]

					if secondCommandValue == "~" {
						err := os.Chdir(home)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}
						continue
					}

					err := os.Chdir(secondCommandValue)
					if err != nil {
						writeCommandOut(err.Error()+"\n", projectPath, liveStart)
						continue
					}
					continue
				} else {
					writeCommandOut("cd args are invalid.\n", projectPath, liveStart)
					continue
				}
			} else if firstCommandName == "live" {
				if len(cmdSplit) == 2 {
					secondCommandValue := cmdSplit[1]
					if secondCommandValue == "stop" {
						liveStart = false
						continue
					} else if secondCommandValue == "status" {
						if liveStart {
							writeCommandOut("live is started.\n", projectPath, false)
						} else {
							writeCommandOut("live is stopped.\n", projectPath, false)
						}
						continue
					} else {
						liveCommandUsage(projectPath)
						continue
					}
				} else if len(cmdSplit) == 3 {
					secondCommandValue := cmdSplit[1]
					thirdCommandValue := cmdSplit[2]
					if secondCommandValue == "init" {
						if liveStart {
							writeCommandOut("live is already started.\n", projectPath, liveStart)
							continue
						}
						absPath, err := filepath.Abs(thirdCommandValue)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}
						if _, err := os.Stat(absPath); !os.IsNotExist(err) {
							writeCommandOut("can't live in the path.\n", projectPath, liveStart)
							continue
						}

						if err := os.Mkdir(absPath, 0751); err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						projectPath = absPath

						r, err := git.PlainInit(projectPath, false)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						liveStart = true

						go watch(r, projectPath, couterHTMLPath)

						err = os.Chdir(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						continue
					} else if secondCommandValue == "start" {
						if liveStart {
							writeCommandOut("live is already started.\n", projectPath, false)
							continue
						}
						absPath, err := filepath.Abs(thirdCommandValue)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						if _, err := os.Stat(absPath); os.IsNotExist(err) {
							writeCommandOut("File doesn't exists\n", projectPath, liveStart)
							continue
						}

						projectPath = absPath

						// fmt.Println(absPath)

						r, err := git.PlainOpen(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						liveStart = true

						go watch(r, projectPath, couterHTMLPath)

						err = os.Chdir(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}
					} else if secondCommandValue == "upload" {
						if liveStart {
							writeCommandOut("you should stop live before.\n", projectPath, false)
							continue
						}

						absPath, err := filepath.Abs(thirdCommandValue)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						if _, err := os.Stat(absPath); os.IsNotExist(err) {
							writeCommandOut("File doesn't exists\n", projectPath, liveStart)
							continue
						}

						projectPath = absPath

						fmt.Println(projectPath)

						// if _, err = os.Stat(gitDirPath); os.IsNotExist(err) {
						// 	writeCommandOut(".git directory not found\n", projectPath, liveStart)
						// 	continue
						// }

						_, err = git.PlainOpen(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						writeCommandOut("preparing to upload ...\n", projectPath, liveStart)

						projectName := filepath.Base(projectPath)

						// gitDirPath := projectName + "/.git"

						prevDir, err := filepath.Abs(".")
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						err = os.Chdir(projectName)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						var buf bytes.Buffer
						err = compress(".git", &buf)
						if err != nil {
							writeCommandOut("compress .git-directory failed\n", projectPath, liveStart)
							continue
						}

						err = os.Chdir(prevDir)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						writeCommandOut("uploading ...\n", projectPath, liveStart)

						// r := bytes.NewReader(buf)

						req, err := http.NewRequest("POST", "https://localhost:3000/api/live/upload?projectName="+projectName, &buf)
						req.Header.Set("Content-Type", "application/gzip")

						client := &http.Client{}
						res, err := client.Do(req)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}
						defer res.Body.Close()

						bodyBytes, err := ioutil.ReadAll(res.Body)
						if err != nil {
							log.Fatal(err)
						}

						if res.StatusCode == http.StatusInternalServerError {
							errorsResponse := ErrorsResponse{}
							err = json.Unmarshal(bodyBytes, &errorsResponse)
							if err != nil {
								writeCommandOut(err.Error()+"\n", projectPath, liveStart)
								continue
							}
							if len(errorsResponse) != 1 {
								writeCommandOut("error response is invalid\n", projectPath, liveStart)
								continue
							}
							writeCommandOut(errorsResponse[0].Message+"\n", projectPath, liveStart)
							continue
						} else if res.StatusCode == http.StatusOK {
							uploadsResponse := UploadsResponse{}
							err = json.Unmarshal(bodyBytes, &uploadsResponse)
							if err != nil {
								writeCommandOut(err.Error()+"\n", projectPath, liveStart)
								continue
							}

							if len(uploadsResponse) != 1 {
								writeCommandOut("upload response is invalid\n", projectPath, liveStart)
								continue
							}
							uploadedURL := uploadsResponse[0].URL
							writeCommandOut("done!\n", projectPath, liveStart)
							writeCommandOut("you can see your live-coding in \""+uploadedURL+"\"\n", projectPath, liveStart)
							continue
						} else {
							writeCommandOut("error code is unknown\n", projectPath, liveStart)
							continue
						}
					} else {
						liveCommandUsage(projectPath)
						continue
					}
				} else {
					liveCommandUsage(projectPath)
					continue
				}
			} else {
				writeCommandInput(line, projectPath, liveStart)
				cmd := exec.Command("bash", "-c", line)
				cmd.Stdout = writer
				cmd.Stderr = writer

				cmd.Run()
				out := buffer.String()
				buffer.Reset()

				writeCommandOut(out, projectPath, liveStart)
			}
		}

	}
}
