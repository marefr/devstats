#!/bin/sh
GHA2DB_PROJECTS_OVERRIDE="+vitess" GHA2DB_LOCAL=1 GHA2DB_PROCESS_COMMITS=1 GHA2DB_PROCESS_REPOS=1 GHA2DB_EXTERNAL_INFO=1 GHA2DB_PROJECTS_COMMITS="vitess" PG_DB=vitess ./get_repos
