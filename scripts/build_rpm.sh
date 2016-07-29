git_branch=$(git symbolic-ref HEAD 2> /dev/null | cut -b 12-)
git_branch="${git_branch:-test}"
git_sha=$(git log --pretty=format:"%h" -1)
git_version=$(git describe | sed -e 's/^v//')
arch="$(uname -m)"

set -e
set -u
set -o pipefail

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__proj_dir="$(dirname "$__dir")"

# shellcheck source=scripts/common.sh
. "${__dir}/common.sh"

_info "project path: ${__proj_dir}"

build_dir="${__proj_dir}/build"
bin_dir="${build_dir}/bin"
plugin_dir="${build_dir}/plugin"

## ubuntu 14.04
_upstart_build=${build_dir}/upstart

mkdir -p ${_upstart_build}/usr/local/bin/
mkdir -p ${_upstart_build}/etc/init
mkdir -p ${_upstart_build}/etc/snapd

cp ${__dir}/initscripts/snapd.conf ${_upstart_build}/etc/init/
cp ${__proj_dir}/examples/configs/snap-config-sample.yaml ${_upstart_build}/etc/snapd/snapd.yaml
cp ${bin_dir}/* ${_upstart_build}/usr/local/bin/

package_name="${_upstart_build}/snapd-${git_version}.el6.${arch}.rpm"
fpm -s dir -t rpm \
  -v ${git_version} -n snapd -a ${arch} --description "Snap is an open telemetry framework designed to simplify the collection, processing and publishing of system data through a single API" \
  --deb-upstart ${__dir}/initscripts/snapd.conf \
  --vendor "Intel Corporation" \
  --license "Apache2.0" -C ${_upstart_build} -p ${package_name} .

## ubuntu 16.04, Debian 8
_systemd_build=${build_dir}/systemd

mkdir -p ${_systemd_build}/usr/local/bin/
mkdir -p ${_systemd_build}/lib/systemd/system/
mkdir -p ${_systemd_build}/etc/snapd

cp ${__proj_dir}/examples/configs/snap-config-sample.yaml ${_systemd_build}/etc/snapd/snapd.yaml
cp ${bin_dir}/* ${_systemd_build}/usr/local/bin/
cp ${__dir}/initscripts/snapd.service $_systemd_build/lib/systemd/system/

package_name="${_systemd_build}/snapd-${git_version}.el7.${arch}.rpm"
fpm -s dir -t rpm \
  -v ${git_version} -n snapd -a ${arch} --description "Snap is an open telemetry framework designed to simplify the collection, processing and publishing of system data through a single API" \
  --config-files /etc/snapd/ \
  --vendor "Intel Corporation" \
  --license "Apache2.0" -C ${_systemd_build} -p ${package_name} .

