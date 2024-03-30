# Migrations
```bash
go install github.com/rubenv/sql-migrate/...@latest
sql-migrate up -env=production -config=internal/infra/config/db/dbconfig.yml
sql-migrate down -env=production -config=internal/infra/config/db/dbconfig.yml
sql-migrate status -env=production -config=internal/infra/config/db/dbconfig.yml
```