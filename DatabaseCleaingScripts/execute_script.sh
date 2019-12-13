
user=$1;
password=$2;
db_name=$3;
script=$4;

PGPASSWORD=$password
PGPASSWORD=$password psql -U $user -d $db_name -f $script;
