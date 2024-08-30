## gRPCについて
### protoファイルの作成
参考: https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/intro
1. `./api`ディレクトリにprotoファイルを作成し記述する。
2. protocコマンドでコードを生成する
```
// backendのコンテナに入る
docker compose exec gabaithon-09-back bash
// apiディレクトリへ移動
cd api
// ファイルを自動生成
protoc --go_out=../pkg/grpc --go_opt=paths=source_relative \
	--go-grpc_out=../pkg/grpc --go-grpc_opt=paths=source_relative \
	{protoファイル名前}
```