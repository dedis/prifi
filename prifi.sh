#!/usr/bin/env bash


# ************************************
# PriFi all-in-one startup script
# ************************************
# author : Ludovic Barman
# email : ludovic.barman@gmail.com
# belongs to : the PriFi project
# 			<github.com/dedis/prifi>
# ************************************

# variables that you might change often

dbg_lvl=3                       # 1=less verbose, 3=more verbose. goes up to 5, but then prints the SDA's message (network framework)
try_use_real_identities="false" # if "true", will try to use "self-generated" public/private key as a replacement for the dummy keys
                                # we generated for you. It asks you if it does not find real keys. If false, will always use the dummy keys.
colors="true"                   # if  "false", the output of PriFi (and this script) will be in black-n-white

socksServer1Port=8080           # the port for the SOCKS-Server-1 (part of the PriFi client)
socksServer2Port=8090           # the port to attempt connect to (from the PriFi relay) for the SOCKS-Server-2
                                # notes : see <https://github.com/dedis/prifi/blob/master/README_architecture.md>

all_localhost_n_clients=3      # number of clients to start in the "all-localhost" script

# default file names :

prifi_file="prifi.toml"                     # default name for the prifi config file (contains prifi-specific settings)
identity_file="identity.toml"               # default name for the identity file (contains public + private key)
group_file="group.toml"                     # default name for the group file (contains public keys + address of other nodes)

# location of the buildable (go build) prifi file :

bin_file="$GOPATH/src/github.com/dedis/prifi/sda/app/prifi.go"

# we have two "identities" directory. The second one is empty unless you generate your own keys with "gen-id"

configdir="config"
defaultIdentitiesDir="identities_default"   # in $configdir
realIdentitiesDir="identities_real"         # in $configdir

sleeptime_between_spawns=1                  # time in second between entities launch in all-localhost part

source "helpers.lib.sh"

# ------------------------
#     HELPER FUNCTIONS
# ------------------------

print_usage() {
	echo
	echo -e "PriFi, a tracking-resistant protocol for local-area anonymity"
	echo
	echo -e "Usage: run-prifi.sh ${highlightOn}role/operation [params]${highlightOff}"
	echo -e "	${highlightOn}role${highlightOff}: client, relay, trustee (interactive); relay-d, trustee-d (same, but in deamon)"
	echo -e "	${highlightOn}operation${highlightOff}: install, all-localhost, gen-id, kill-d"
	echo -e "	${highlightOn}params${highlightOff} for role ${highlightOn}relay${highlightOff}, ${highlightOn}relay-d${highlightOff}: [socks_server_port] (optional, numeric)"
	echo -e "	${highlightOn}params${highlightOff} for role ${highlightOn}trustee${highlightOff}, ${highlightOn}trustee-d${highlightOff}: id (required, numeric)"
	echo -e "	${highlightOn}params${highlightOff} for role ${highlightOn}client${highlightOff}: id (required, numeric), [prifi_socks_server_port] (optional, numeric)"
	echo -e "	${highlightOn}params${highlightOff} for operation ${highlightOn}install${highlightOff}: none"
	echo -e "	${highlightOn}params${highlightOff} for operation ${highlightOn}all-localhost${highlightOff}: none"
	echo -e "	${highlightOn}params${highlightOff} for operation ${highlightOn}gen-id${highlightOff}: none"
	echo -e "	${highlightOn}params${highlightOff} for operation ${highlightOn}kill-d${highlightOff}: none"
	echo

	echo -e "Man-page:"
	echo -e "	${highlightOn}install${highlightOff}: get the dependencies, and tests the setup"
	echo -e "	${highlightOn}relay${highlightOff}: starts a PriFi relay"
	echo -e "	${highlightOn}relay-d${highlightOff}: starts a PriFi relay in background"
	echo -e "	${highlightOn}trustee${highlightOff}: starts a PriFi trustee, using the config file trustee${highlightOn}id${highlightOff}"
	echo -e "	${highlightOn}trustee-d${highlightOff}: starts a PriFi trustee, using the config file trustee${highlightOn}id${highlightOff} in background"
	echo -e "	${highlightOn}client${highlightOff}: starts a PriFi client, using the config file client${highlightOn}id${highlightOff}"
	echo -e "	${highlightOn}all-localhost${highlightOff}: starts a Prifi relay, a trustee, three clients all on localhost"
	echo -e "	${highlightOn}gen-id${highlightOff}: interactive creation of identity.toml"
	echo -e "	${highlightOn}kill-d${highlightOff}: kills all prifi-deamons (relay-d and trustee-d's)"
	echo -e "	Lost ? read https://github.com/dedis/prifi/README.md"
}

# ------------------------
#     MAIN SWITCH
# ------------------------

# $1 is operation : "install", "relay", "client", "trustee", "sockstest", "all-localhost", "clean", "gen-id"
case $1 in

	install|Install|INSTALL)

		echo -n "Testing for GO... "
		test_go
		echo -e "$okMsg"

		echo -n "Getting all go packages... "
		cd sda/app; go get -u -v  ./...
		cd ../..
		cd socks; go get -u -v  ./...
		cd ..
		cd prifi-lib; go get -u -v  ./...
		cd ..
		cd prifi-mobile; go get -u -v ./...
		cd ..

		# this package somehow fails to install itself
		go get -u -v github.com/montanaflynn/stats

		echo -e "$okMsg"
		;;

	relay|Relay|RELAY)

		#test for proper setup
		test_go

		# the 2nd argument can replace the port number
		if [ "$#" -eq 2 ]; then
			test_digit "$2" 2
			socksServer2Port="$2"
		fi

		#specialize the config file (we use the dummy folder, and maybe we replace with the real folder after)
		prifi_file2="$configdir/$prifi_file"
		identity_file2="$configdir/$defaultIdentitiesDir/relay/$identity_file"
		group_file2="$configdir/$defaultIdentitiesDir/relay/$group_file"

		#we we want to, try to replace with the real folder
		if [ "$try_use_real_identities" = "true" ]; then
			if [ -f "$configdir/$realIdentitiesDir/relay/$identity_file" ] && [ -f "$configdir/$defaultIdentitiesDir/relay/$group_file" ]; then
				echo -e "$okMsg Found real identities (in $configdir/$realIdentitiesDir/relay/), using those."
				identity_file2="$configdir/$realIdentitiesDir/relay/$identity_file"
				group_file2="$configdir/$realIdentitiesDir/relay/$group_file"
			else
				echo -e "$warningMsg Trying to use real identities, but does not exists for relay (in $configdir/$realIdentitiesDir/relay/). Falling back to pre-generated ones."
			fi
		else
			echo -e "$warningMsg using pre-created identities. Set \"try_use_real_identities\" to True in real deployements."
		fi

		# test that all files exists
		test_files

		#run PriFi in relay mode
		DEBUG_COLOR="$colors" go run "$bin_file" --cothority_config "$identity_file2" --group "$group_file2" -d "$dbg_lvl" --prifi_config "$prifi_file2" --port "$socksServer1Port" --port_client "$socksServer2Port" relay
		;;

	trustee|Trustee|TRUSTEE)

		trusteeId="$2"

		#test for proper setup
		test_go

		if [ "$#" -lt 2 ]; then
			echo -e "$errorMsg parameter 2 need to be the trustee id."
			exit 1
		fi
		test_digit "$trusteeId" 2

		#specialize the config file (we use the dummy folder, and maybe we replace with the real folder after)
		prifi_file2="$configdir/$prifi_file"
		identity_file2="$configdir/$defaultIdentitiesDir/trustee$trusteeId/$identity_file"
		group_file2="$configdir/$defaultIdentitiesDir/trustee$trusteeId/$group_file"

		#we we want to, try to replace with the real folder
		if [ "$try_use_real_identities" = "true" ]; then
			if [ -f "$configdir/$realIdentitiesDir/trustee$trusteeId/$identity_file" ] && [ -f "$configdir/$defaultIdentitiesDir/trustee$trusteeId/$group_file" ]; then
				echo -e "$okMsg Found real identities (in $configdir/$realIdentitiesDir/trustee$trusteeId/), using those."
				identity_file2="$configdir/$realIdentitiesDir/trustee$trusteeId/$identity_file"
				group_file2="$configdir/$realIdentitiesDir/trustee$trusteeId/$group_file"
			else
				echo -e "$warningMsg Trying to use real identities, but does not exists for trustee $trusteeId (in $configdir/$realIdentitiesDir/trustee$trusteeId/). Falling back to pre-generated ones."
			fi
		else
			echo -e "$warningMsg using pre-created identities. Set \"try_use_real_identities\" to True in real deployements."
		fi

		# test that all files exists
		test_files

		#run PriFi in relay mode
		DEBUG_COLOR="$colors" go run "$bin_file" --cothority_config "$identity_file2" --group "$group_file2" -d "$dbg_lvl" --prifi_config "$prifi_file2" --port "$socksServer1Port" --port_client "$socksServer2Port" trustee
		;;

	client|Client|CLIENT)

		clientId="$2"

		#test for proper setup
		test_go

		if [ "$#" -lt 2 ]; then
			echo -e "$errorMsg parameter 2 need to be the client id."
			exit 1
		fi
		test_digit "$clientId" 2

		# the 3rd argument can replace the port number
		if [ "$#" -eq 3 ]; then
			test_digit "$3" 3
			socksServer1Port="$3"
		fi

		#specialize the config file (we use the dummy folder, and maybe we replace with the real folder after)
		prifi_file2="$configdir/$prifi_file"
		identity_file2="$configdir/$defaultIdentitiesDir/client$clientId/$identity_file"
		group_file2="$configdir/$defaultIdentitiesDir/client$clientId/$group_file"

		#we we want to, try to replace with the real folder
		if [ "$try_use_real_identities" = "true" ]; then
			if [ -f "$configdir/$realIdentitiesDir/client$clientId/$identity_file" ] && [ -f "$configdir/$realIdentitiesDir/client$clientId/$group_file" ]; then
				echo -e "$okMsg Found real identities (in $configdir/$realIdentitiesDir/client$clientId/), using those."
				identity_file2="$configdir/$realIdentitiesDir/client$clientId/$identity_file"
				group_file2="$configdir/$realIdentitiesDir/client$clientId/$group_file"
			else
				echo -e "$warningMsg Trying to use real identities, but does not exists for client $clientId (in $configdir/$realIdentitiesDir/client$clientId/). Falling back to pre-generated ones."
			fi
		else
			echo -e "$warningMsg using pre-created identities. Set \"try_use_real_identities\" to True in real deployements."
		fi

		# test that all files exists
		test_files

		#run PriFi in relay mode
		DEBUG_COLOR="$colors" go run "$bin_file" --cothority_config "$identity_file2" --group "$group_file2" -d "$dbg_lvl" --prifi_config "$prifi_file2" --port "$socksServer1Port" --port_client "$socksServer2Port" client
		;;

	localhost|Localhost|LOCALHOST|all-localhost|All-Localhost|ALL-LOCALHOST)

		thisScript="$0"
		if [ "$try_use_real_identities" = "true" ]; then
			echo -en "$warningMsg, try_use_real_identities set to true, but this is incompatible to all-localhost mode. Switching to false ..."
			sed -i -e 's/try_use_real_identities=\"true\"/try_use_real_identities=\"false\"/g' "$thisScript"
			echo -e "$okMsg"
		fi

		#test for proper setup
		test_go

		#test if a socks proxy is already running (needed for relay), or start ours
		socks=$(netstat -tunpl 2>/dev/null | grep "$socksServer2Port" | wc -l)

		if [ "$socks" -ne 1 ]; then
			echo -n "Socks proxy not running, starting it... "
			cd socks && ./run-socks-proxy.sh "$socksServer2Port" > ../socks.log 2>&1 &
			SOCKSPID=$!
			echo -e "$okMsg"
		fi

		echo -n "Starting relay...			"
		"$thisScript" relay > relay.log 2>&1 &
		RELAYPID=$!
		THISPGID=$(ps -o pgid= "$RELAYPID" | sed -e "s/^ //")

		trap 'echo; echo "Gonna run kill -TERM -- -$THISPGID"; kill -TERM -- -$THISPGID' 2

		echo -e "$okMsg"

		sleep "$sleeptime_between_spawns"

		echo -n "Starting trustee 0...			"
		"$thisScript" trustee 0 > trustee0.log 2>&1 &
		echo -e "$okMsg"

		sleep "$sleeptime_between_spawns"

		for i in `seq 0 $(($all_localhost_n_clients - 1))`
		do
			port=$(($socksServer1Port + $i))
			log="client$i.log"
			echo -n "Starting client $i... (SOCKS on :$port)	"
			"$thisScript" client "$i" "$port" > "$log" 2>&1 &
			echo -e "$okMsg"
			sleep "$sleeptime_between_spawns"

		done

		echo "PriFi deployed. Press ^C to kill all..."
		wait
		;;

	gen-id|Gen-Id|GEN-ID)
		echo -e "Going to generate private/public keys (named ${highlightOn}identity.toml${highlightOff})..."

		read -p "Do you want to generate it for [r]elay, [c]lient, or [t]trustee ? " key

		path=""
		case "$key" in
			r|R)
				path="relay"
			;;
			t|T)

				read -p "Do you want to generate it for trustee [0] or [1] (or more - enter digit) ? " key2

				test_digit "$key2" 1
				pathSource="trustee0"
				path="trustee$key2"
				;;

			c|C)
				read -p "Do you want to generate it for client [0],[1] or [2] (or more - enter digit) ? " key2

				test_digit "$key2" 1
				pathSource="client0"
				path="client$key2"
				;;


			*)
				echo -e "$errorMsg did not understand."
				exit 1
				;;
		esac

		pathReal="$configdir/$realIdentitiesDir/$path/"
		pathDefault="$configdir/$defaultIdentitiesDir/$pathSource/"
		echo -e "Gonna generate ${highlightOn}identity.toml${highlightOff} in ${highlightOn}$pathReal${highlightOff}"

		#generate identity.toml
		DEBUG_COLOR="$colors" go run "$bin_file" --default_path "$pathReal" gen-id

		if [ ! -f "${pathReal}group.toml" ]; then
			#now group.toml
			echo -n "Done ! now copying group.toml from identities_default/ to identity_real/..."
			cp "${pathDefault}/group.toml" "${pathReal}group.toml"
			echo -e "$okMsg"

			echo -e "Please edit ${highlightOn}$pathReal/group.toml${highlightOff} to the correct values."
		else
			echo -e "Group file ${highlightOn}$pathReal/group.toml${highlightOff} already exists, not overwriting! you might want to check that the contents are correct."
		fi
		;;

	relay-d)

		#test for proper setup
		test_go

		#if running, kill the previous one
		if [ -f .GPID_RELAY ]; then
			gpid=$(cat .GPID_RELAY)
			echo -e "PriFi relay already running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_RELAY
		fi

		thisScript="$0"

		echo -n "Starting relay...			"
		"$thisScript" relay > relay.log 2>&1 &
		RELAYPID=$!
		RELAYPGID=$(ps -o pgid= "$RELAYPID")
		echo -e "$okMsg"

		echo -e "PriFi relay deployed, PGID $RELAYPGID. Kill with \"kill -TERM -- -$RELAYPGID\""

		rm -f .GPID_RELAY
		echo $RELAYPGID > .GPID_RELAY
		;;

	trustee-d)

		#test for proper setup
		test_go

		#if running, kill the previous one
		if [ -f .GPID_TRUSTEE ]; then
			gpid=$(cat .GPID_TRUSTEE)
			echo -e "PriFi trustee already running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_TRUSTEE
		fi

		thisScript="$0"
		trusteeId="$2"

		if [ "$#" -lt 2 ]; then
			echo -e "$errorMsg parameter 2 need to be the trustee id."
			exit 1
		fi
		test_digit "$trusteeId" 2

		echo -n "Starting trustee $trusteeId...			"
		"$thisScript" trustee "$trusteeId" > trustee${trusteeId}.log 2>&1 &
		TRUSTEEPID=$!
		TRUSTEEGPID=$(ps -o pgid= "$TRUSTEEPID")
		echo -e "$okMsg"

		echo -e "PriFi trustee deployed, PGID $TRUSTEEGPID. Kill with \"kill -TERM -- -$TRUSTEEGPID\""

		rm -f .GPID_TRUSTEE
		echo $TRUSTEEGPID > .GPID_TRUSTEE
		;;

	socks-d)

		#if running, kill the previous one
		if [ -f .GPID_SOCKSEXIT ]; then
			gpid=$(cat .GPID_SOCKSEXIT)
			echo -e "PriFi socks exit already running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_SOCKSEXIT
		fi

		echo -n "Starting SOCKS Server...			"
		cd socks && ./run-socks-proxy.sh "$socksServer2Port" > ../socks.log 2>&1 &
		SOCKSPID=$!
		SOCKSPGID=$(ps -o pgid= "$SOCKSPID")
		echo -e "$okMsg"

		echo -e "PriFi socks exit deployed, PGID $SOCKSPGID. Kill with \"kill -TERM -- -$SOCKSPGID\""

		rm -f .GPID_SOCKSEXIT
		echo $SOCKSPGID > .GPID_SOCKSEXIT
		;;

	kill-d)

		if [ -f .GPID_RELAY ]; then
			gpid=$(cat .GPID_RELAY)
			echo -e "PriFi relay running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_RELAY
		else
			echo -e "PriFi relay not running, ignoring"
		fi

		if [ -f .GPID_TRUSTEE ]; then
			gpid=$(cat .GPID_TRUSTEE)
			echo -e "PriFi trustee running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_TRUSTEE
		else
			echo -e "PriFi trustee not running, ignoring"
		fi
		
		if [ -f .GPID_SOCKSEXIT ]; then
			gpid=$(cat .GPID_SOCKSEXIT)
			echo -e "PriFi socks exit running, killing old instance with kill -TERM -- -$gpid"			
			kill -TERM -- -"$gpid"
			rm -f .GPID_SOCKSEXIT
		else
			echo -e "PriFi socks-exit not running, ignoring"
		fi

		;;

	web-server)

		cd web
		go run index.go &
		;;


	clean|Clean|CLEAN)
		echo -n "Cleaning local log files... 			"
		rm *.log 1>/dev/null 2>&1
		echo -e "$okMsg"
		;;

	*)
		print_usage
		;;
esac
