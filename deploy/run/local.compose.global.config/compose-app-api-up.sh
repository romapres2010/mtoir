export ENV_FILE=$(pwd)/.env
echo "Environment file:" $ENV_FILE
export LOG_FILE=$(pwd)/log/compose-app-api-up.log
echo "Log file:" $LOG_FILE
(cd .. && ./_script/compose-app-up.sh $ENV_FILE app-api 1>$LOG_FILE 2>&1)
