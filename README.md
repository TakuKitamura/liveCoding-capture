# liveCoding-capture
Version 0.1

## work
### 円周率を計算している例

Webサイトを開いたら､一番下のバーを動かしてみてください｡動画に表示されているIDと対応しています｡

動画 https://drive.google.com/file/d/1Nw1GjTHxQYJ7kNT79moPbDwlltI6Z0gI/view?usp=sharing

Webサイト https://live-coding.takukitamura.com/?id=qikdVYiv1PeX7nL09DkV

### 【自作言語】スイーツ絵文字で円周率計算してみた Part1 文字列出力編

動画 https://www.youtube.com/watch?v=llXQKdGGk7M

Webサイト(読み込みが少し遅いです) https://live-coding.takukitamura.com/?id=dj9lWEZK9BiSvYmDf0gA

## 動作確認済み
- Mac
- Amaozon Linux

## dependency
- Golang
- dep (go dependency management tool)
- git (いずれ必要なくなるはずです｡)

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
