#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Required env vars
# Check that env vars are set
set -x
for v in TO_HOST TO_PORT TO_ADMIN_USER TO_ADMIN_PASSWORD; do
    [[ -z $(eval echo \$$v) ]] || continue
    echo "$v is unset"
    exit 1
done

. /to-access.sh

TO_URL="https://$TO_HOST:$TO_PORT"
# wait until the ping endpoint succeeds
while ! to-get api/1.3/ping 2>/dev/null; do
   echo waiting for trafficops
   sleep 3
done

# NOTE: order dependent on foreign key references, e.g. tenant must be defined before user
endpoints="tenant user cdn server"

load_data_from() {
    local dir="$1"
    if [[ ! -d $dir ]] ; then
        echo "Failed to load data from '$dir': directory does not exist"
    fi

    for ep in $endpoints; do
        d="$dir/$ep"
        [[ -d $d ]] || continue
        echo "Loading data from $d"
        for f in "$d"/*.json; do
            [[ -r $f ]] || continue
            t=$(mktemp --tmpdir $ep-XXX.json)
            envsubst <"$f" >"$t"
            to-post api/1.3/"$ep" "$t"
            rm "$t"
        done
    done
}

# First,  load required data at the top level
load_data_from /traffic_ops_data

# If TO_DATA is defined, load from subdirs with that name (space-separated)
if [[ -n $TO_DATA ]]; then
    for subdir in $TO_DATA; do
        load_data_from /traffic_ops_data/$subdir
    done
fi


