while true
do
  date
  # GHA2DB_ST=1 GHA2DB_DEBUG=1 GHA2DB_QOUT=1 ./sync_ruby.sh
  ./sync_ruby.sh
  date
  sleep $1
done
