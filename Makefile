all: setup migrate

setup:
	dropdb henrymed || true
	createdb henrymed --owner=postgres
	cd migrations && goose postgres "user=postgres dbname=henrymed sslmode=disable" up

create:
	cd migrations && goose create $(name) sql

upgrade-local:
	goose -dir migrations postgres "user=postgres dbname=henrymed sslmode=disable" up

downgrade-local:
	goose -dir migrations postgres "user=postgres dbname=henrymed sslmode=disable" down