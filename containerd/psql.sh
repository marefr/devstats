#!/bin/bash
set -o pipefail
> errors.txt
> run.log
GHA2DB_PROJECT=containerd IDB_DB=containerd PG_DB=containerd GHA2DB_LOCAL=1 ./structure 2>>errors.txt | tee -a run.log || exit 1
GHA2DB_PROJECT=containerd IDB_DB=containerd PG_DB=containerd GHA2DB_LOCAL=1 ./gha2db 2015-12-17 0 today now 'containerd,docker/containerd' 2>>errors.txt | tee -a run.log || exit 2
GHA2DB_PROJECT=containerd IDB_DB=containerd PG_DB=containerd GHA2DB_LOCAL=1 GHA2DB_MGETC=y GHA2DB_SKIPTABLE=1 GHA2DB_INDEX=1 ./structure 2>>errors.txt | tee -a run.log || exit 3
./grafana/influxdb_recreate.sh containerd
./containerd/setup_repo_groups.sh 2>>errors.txt | tee -a run.log || exit 4
./containerd/import_affs.sh 2>>errors.txt | tee -a run.log || exit 5
./containerd/setup_scripts.sh 2>>errors.txt | tee -a run.log || exit 6
./containerd/get_repos.sh 2>>errors.txt | tee -a run.log || exit 7
echo "All done. You should run ./containerd/reinit.sh script now."
