SET APP_CONFIG_FILE=.\..\..\cfg\app.global.yaml
SET APP_META_ENTITY_CONFIG_FILE=.\..\..\cfg\entities\global_entities.yaml
SET APP_HTTP_LISTEN_SPEC=127.0.0.1:3000
SET APP_LOG_LEVEL=DEBUG
SET APP_LOG_FILE=.\log\app.log
SET APP_PG_USER=postgres
SET APP_PG_PASS=postgres
SET APP_PG_HOST=localhost
SET APP_PG_PORT=5437
SET APP_PG_DBNAME=postgres

.\app.exe

pause