#!bin/bash

PID=$1
max=0
while true    
do
    mem_use=$(cat /proc/$1/status | grep RSS | grep -o '[0-9]\+')
    if [[ $mem_use -gt $max ]];then
        max=$mem_use
    fi

    #cat /proc/$1/status | grep RSS | grep -o '[0-9]\+' >> /tmp/m.txt
    sleep 0.01
    check=$(ps --no-heading $1 | wc -l)
    if [ $check != "1" ];then
        break
    fi
done
echo $max > /tmp/m.txt

