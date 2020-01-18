# liveCoding-capture
Version 0.0.1
もうすぐAPIを公開する予定｡

## dependency
- Golang
- dep (go dependency management tool)

## install
```sh
$ pwd
liveCoding-capture
$ git clone https://github.com/TakuKitamura/liveCoding-capture.git
$ dep ensure
$ go run liveCodingCapture.go
Welcome Live Coding Capture! (v0.0.1)
Please open "xxx.html" in your browser.
(stopped) $
```

## embedded commands
```
$ live init (ProjectPath) # initialize project and start capture
$ live status # check live status
$ live start (ProjectPath) # start capture
$ live stop # stop live
$ live upload # your live-coding is shared on the internet 
```