#!/usr/bin/env bash
# not executable ! use prifi.sh, simul.sh, test.sh

# min required go version
min_go_version=19                           # min required go version, without the '.', e.g. 17 for 1.7.x

#pretty colored message
highlightOn="\033[33m"
highlightOff="\033[0m"
shell="\033[35m[script]${highlightOff}"
warningMsg="${highlightOn}[warning]${highlightOff}"
errorMsg="\033[31m\033[1m[error]${highlightOff}"
okMsg="\033[32m[ok]${highlightOff}"
if [ "$colors" = "false" ]; then
	highlightOn=""
	highlightOff=""
	shell="[script]"
	warningMsg="[warning]"
	errorMsg="[error]"
	okMsg="[ok]"
fi

#tests if GOPATH is set and exists
test_go(){
    if [ -z "$GOPATH"  ]; then
        echo -e "$errorMsg GOPATH is unset ! make sure you installed the Go language."
        exit 1
    fi
    if [ ! -d "$GOPATH"  ]; then
        echo -e "$errorMsg GOPATH ($GOPATH) is not a folder ! make sure you installed the Go language correctly."
        exit 1
    fi
    GO_VER=$(go version 2>&1 | sed 's/.*version go\([[:digit:]]*\)\.\([[:digit:]]*\)\(.*\)/\1\2/; 1q')
    if [ "$GO_VER" -lt "$min_go_version" ]; then
        echo "$errorMsg Go >= $(echo $min_go_version | sed 's/./&./') is required"
        exit 1
    fi
}

# test if $1 is a digit, if not, prints "argument $2 invalid" and exit.
test_digit() {
    case $1 in
        ''|*[!0-9]*)
            echo -e "$errorMsg parameter $2 need to be an integer."
            exit 1;;
        *) ;;
    esac
}

#test if all the files we need are there.
test_files() {

	if [ ! -f "$bin_file" ]; then
		echo -e "$errorMsg Runnable go file does not seems to exists: $bin_file"
		exit
	fi

	if [ ! -f "$identity_file2" ]; then
		echo -e "$errorMsg Cothority config file does not exist: $identity_file2"
		exit
	fi

	if [ ! -f "$group_file2" ]; then
		echo -e "$errorMsg Cothority group file does not exist: $group_file2"
		exit
	fi

	if [ ! -f "$prifi_file2" ]; then
		echo -e "$errorMsg PriFi config file does not exist: $prifi_file2"
		exit
	fi
}
