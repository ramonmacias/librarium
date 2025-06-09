# Librarium (WORK IN PROGRES)

In order to handle all the database migrations in Postgres, we used the well known project golang-migrate, please follow the instructions in here https://github.com/golang-migrate/migrate/tree/master/cmd/migrate so you can install the CLI

$ migrate -source file://internal/postgres/migration -database 'postgres://librarium_user:librarium_pass@localhost:5432/librarium sslmode=disable' up

$ migrate create -dir internal/postgres/migration -ext sql librarium