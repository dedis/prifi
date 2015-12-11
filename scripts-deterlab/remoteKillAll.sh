#!/usr/local/bin/bash

source ~/config.sh

#max trustee minus one, really
maxtrustee=4
maxclient=9

echo "Remoting inside relay"
ssh router.LB-LLD.SAFER.isi.deterlab.net "pkill -f ${program}"
ssh router.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${nohupoutfolder}${nohuprelayname}${nohupext}"
ssh router.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${logPath}relay.out"

# Start clients
for i in $(seq 0 $maxclient); do
  echo "Remoting inside client-$i"
  ssh client-$i.LB-LLD.SAFER.isi.deterlab.net "pkill -f ${program}"
  ssh client-$i.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${nohupoutfolder}${nohupclientname}${id}${nohupext}"
  ssh client-$i.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${logPath}client${i}.out"
done

# Start trustees
for i in $(seq 0 $maxtrustee); do
  echo "Remoting inside trustee-$i"
  ssh trustee-$i.LB-LLD.SAFER.isi.deterlab.net "pkill -f ${program}"
  ssh trustee-$i.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${nohupoutfolder}${nohuprelayname}${nohupext}"
  ssh trustee-$i.LB-LLD.SAFER.isi.deterlab.net "rm -rf ${logPath}trusteeServer.log"
done