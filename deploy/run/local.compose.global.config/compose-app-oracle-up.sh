export ENV_FILE=$(pwd)/.env
echo "Environment file:" $ENV_FILE
export LOG_FILE=$(pwd)/log/compose-app-oracle-up.log
echo "Log file:" $LOG_FILE
(cd .. && ./_script/compose-app-up.sh $ENV_FILE app-oracle 1>$LOG_FILE 2>&1)
