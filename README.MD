# GoFrame Template For SingleRepo

Quick Start: 
- https://goframe.org/pages/viewpage.action?pageId=1114399


```bash
go run main.go
```

```bash
go run balance/main.go
```

### DB migration

- Install [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) CLI. Read [more](https://www.freecodecamp.org/news/database-migration-golang-migrate/)
- Create Migration

  ```bash
  migrate create -ext sql -dir manifest/migrations/ -seq {migration_name}
  ```

- Migration Up

  ```bash
  export POSTGRESQL_URL='postgres://postgres:password@localhost:5432/postgres?sslmode=disable'

  migrate -database ${POSTGRESQL_URL} -path migrate -database ${POSTGRESQL_URL} -path manifest/migrations up
  ```

- Migration Down

  ```bash
  migrate -database ${POSTGRESQL_URL} -path migrate -database ${POSTGRESQL_URL} -path manifest/migrations down
  ```

- Migration Version

  ```bash
  migrate -database ${POSTGRESQL_URL} -path manifest/migrations version
  ```

- Migration Fix

  ```bash
  migrate -database ${POSTGRESQL_URL} -path manifest/migrations force {version}
  ```