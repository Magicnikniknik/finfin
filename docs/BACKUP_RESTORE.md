# Backup / Restore

## Backup PostgreSQL

```bash
docker exec -t finfin-postgres pg_dump -U postgres -d finfin > finfin_backup.sql
```

## Restore PostgreSQL

```bash
cat finfin_backup.sql | docker exec -i finfin-postgres psql -U postgres -d finfin
```

## Important

Back up before:

- upgrades
- migration changes
- seed resets in pilot/demo environments

## Persistent data

Postgres data is stored in the `finfin_postgres_data` volume.
