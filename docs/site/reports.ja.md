# Reports 日本語版

[English](reports.md)

reports は、Mockport の test run が何を実行し、どの safety check に引っかかったかを確認するための user-facing evidence です。

## 確認方法

- HTTP では `/_mockport/report` を参照します。
- CLI では `mockport report` または JSON format を使います。
- CI では report を artifact として保存すると、adapter coverage と safety status を追いやすくなります。
