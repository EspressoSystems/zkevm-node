#!/bin/sh

set -e

gen() {
    local package=$1

    abigen --bin bin/${package}.bin --abi abi/${package}.abi --pkg=${package} --out=${package}/${package}.go
}

gen_iface() {
    local package=$1

    abigen --abi abi/${package}.abi --pkg=${package} --out=${package}/${package}.go
}

gen polygonzkevm
gen polygonzkevmbridge
gen matic
gen polygonzkevmglobalexitroot
gen mockverifier
gen ihotshot
