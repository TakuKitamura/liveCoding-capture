package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Commit struct {
	ProjectPath string `bson:"project_path"`
	ProjectName string `bson:"project_name"`
	projectPath string `bson:"project_path"`
	Hash        string `bson:"hash"`
	Time        int64  `bson:"time"`
	ID          int    `bson:"id"`
	// Files       map[string]string `bson:"files"`
}

var liveStart bool

const CUI_LOG = ".cui.log"
const LIVE_CODING_PATH = "/Users/kitamurataku/work/liveCoding"

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

func watch(r *git.Repository, projectPath string) error {
	i := 0
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	for {
		if liveStart == false {
			return nil
		}

		time.Sleep(time.Second * 1)

		// TODO: 対応があれば置き換え｡
		// _, err := w.Add(".")
		// CheckIfError(err)

		// CheckIfError(err)

		status, err := w.Status()
		if err != nil {
			continue
		}
		// fmt.Println(status.String())

		if len(status) != 0 {
			cmd := exec.Command("git", "add", projectPath)
			cmd.Dir = w.Filesystem.Root()
			err = cmd.Run()
			if err != nil {
				continue
			}

			commit, err := w.Commit(strconv.FormatInt(time.Now().UnixNano(), 10), &git.CommitOptions{
				Author: &object.Signature{
					When: time.Now(),
				},
			})
			if err != nil {
				continue
			}

			w.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName("master"),
			})

			obj, err := r.CommitObject(commit)
			if err != nil {
				continue
			}

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
			if err != nil {
				continue
			}
			defer client.Disconnect(ctx)

			commitTime, err := strconv.ParseInt(obj.Message, 10, 64)
			if err != nil {
				continue
			}

			commitStruct := Commit{
				ProjectPath: projectPath,
				ProjectName: "test",
				Hash:        obj.Hash.String(),
				Time:        commitTime,
				ID:          i,
				// Files:       files,
			}
			commit_collection := client.Database("liveCoding").Collection("commit")
			_, err = commit_collection.InsertOne(ctx, commitStruct)
			if err != nil {
				continue
			}

			i++

		}
	}
}

func main() {
	projectPath := ""
	buffer := &bytes.Buffer{}
	writer := buffer
	liveStart = false

	fmt.Println("\x1b[32mWelcome Live Coding Capture! (v0.0.1)\x1b[0m")

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
					if secondCommandValue == "start" {
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

						go watch(r, projectPath)

						err = os.Chdir(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
							continue
						}

						continue
					} else if secondCommandValue == "resume" {
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

						go watch(r, projectPath)

						err = os.Chdir(projectPath)
						if err != nil {
							writeCommandOut(err.Error()+"\n", projectPath, liveStart)
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
