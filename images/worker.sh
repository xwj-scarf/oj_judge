#!/bin/bash

#chmod 777 /tmp/input.txt
#chmod 777 code

begin_time=$(date +%s%N)

./tmp/code << EOF > /tmp/output.txt 
   ` PID=$(ps -ef | grep worker.sh | grep -v grep | sed -n '3p'|awk '{print $2}')
     sh /tmp/check_memory.sh $PID >& /tmp/tmp.txt &
     while read LINE
     do
        echo $LINE
     done < /tmp/input.txt      
        `
EOF

end_time=$(date +%s%N)
dif=$((end_time - begin_time))
use_time=$((dif / 1000000))
echo $use_time > /tmp/time.txt

