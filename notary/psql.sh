#!/bin/bash
set -o pipefail
> errors.txt
> run.log
GHA2DB_PROJECT=notary IDB_DB=notary PG_DB=notary GHA2DB_LOCAL=1 ./structure 2>>errors.txt | tee -a run.log || exit 1
GHA2DB_PROJECT=notary IDB_DB=notary PG_DB=notary GHA2DB_LOCAL=1 ./gha2db 2015-06-22 0 today now 'theupdateframework,docker' 'notary' 2>>errors.txt | tee -a run.log || exit 2
GHA2DB_PROJECT=notary IDB_DB=notary PG_DB=notary GHA2DB_LOCAL=1 GHA2DB_MGETC=y GHA2DB_SKIPTABLE=1 GHA2DB_INDEX=1 ./structure 2>>errors.txt | tee -a run.log || exit 3
./grafana/influxdb_recreate.sh notary
./notary/setup_repo_groups.sh 2>>errors.txt | tee -a run.log || exit 4
./notary/import_affs.sh 2>>errors.txt | tee -a run.log || exit 5
./notary/setup_scripts.sh 2>>errors.txt | tee -a run.log || exit 6
./notary/get_repos.sh 2>>errors.txt | tee -a run.log || exit 7
echo "All done. You should run ./notary/reinit.sh script now."
