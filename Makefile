postgres:
	docker run --name simple_bank_db -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -p 5432:5432 -d postgres:12-alpine
createdb:
	docker exec -it simple_bank_db createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it simple_bank_db dropdb simple_bank

# migration commands use golang-migrate tool
migrateup:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrateup1:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simple_bank?sslmode=disable" -verbose up 1
	
migratedown1:
	migrate -path db/migration -database "postgresql://root:root@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...
server:
	go run main.go

mock:
	mockgen \
	--build_flags=--mod=mod \
	-destination=db/mock/store.go \
	-package=mockdb \
	github.com/tiendat0139/simple-bank/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb \
		--grpc-gateway_opt paths=source_relative \
		proto/*.proto
redis:
	docker run --name redis -p 6379:6379 -d redis:7.4.4-alpine

evans:
	evans --host localhost --port 9090 -r repl
.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock proto evans redis
