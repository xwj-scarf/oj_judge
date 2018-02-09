#!/bin/bash

chmod 777 /tmp/input.txt
chmod 777 code

begin_time=$(date +%s%N)
(while read LINE
do 
./code << EOF 
    $LINE 
EOF
done  < /tmp/input.txt
) > /tmp/output.txt

end_time=$(date +%s%N)
dif=$((end_time - begin_time))
use_time=$((dif / 1000000))
echo $use_time

