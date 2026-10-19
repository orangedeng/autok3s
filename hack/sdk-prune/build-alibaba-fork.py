#!/usr/bin/env python3
"""Build a pruned fork of the alibaba cloud SDK for autok3s.

Usage:
    hack/sdk-prune/build-alibaba-fork.py <dst-dir> [<src-module-dir>]

<src-module-dir> defaults to the version pinned in go.mod from the local
module cache. The result keeps the core sdk/ package plus only the API
files autok3s actually calls (services/ecs, services/vpc); everything
else from the ~90 MB upstream module is dropped.

Keep the API lists below in sync with pkg/providers/alibaba and
re-run this script when bumping the SDK.
"""

import os
import re
import shutil
import subprocess
import sys

DST = sys.argv[1].rstrip('/') + '/' if len(sys.argv) > 1 else './thirdparty/alibaba/'

ecs_apis = ['AuthorizeSecurityGroup', 'CreateSecurityGroup', 'DeleteInstances',
            'DescribeInstanceStatus', 'DescribeInstances', 'DescribeSecurityGroupAttribute',
            'DescribeSecurityGroups', 'DescribeVSwitches', 'DescribeVpcs',
            'ListTagResources', 'RunInstances']
vpc_apis = ['AllocateEipAddress', 'AssociateEipAddress', 'CreateVSwitch', 'CreateVpc',
            'DescribeEipAddresses', 'DescribeVSwitchAttributes', 'DescribeVSwitches',
            'DescribeVpcAttribute', 'DescribeVpcs', 'ListTagResources',
            'ReleaseEipAddress', 'TagResources', 'UnassociateEipAddress']

if len(sys.argv) > 2:
    SRC = sys.argv[2].rstrip('/') + '/'
else:
    MC = subprocess.run(['go', 'env', 'GOMODCACHE'], capture_output=True, text=True).stdout.strip()
    SRC = MC + '/github.com/aliyun/alibaba-cloud-sdk-go@v1.61.381/'


def find_api_file(svc, api):
    for fn in os.listdir(SRC + f'services/{svc}'):
        if fn == 'client.go':
            continue
        head = open(SRC + f'services/{svc}/{fn}', errors='ignore').read(200000)
        if f'type {api}Request struct' in head:
            return fn
    return None


def main():
    shutil.rmtree(DST, ignore_errors=True)
    os.makedirs(DST + 'services/ecs')
    os.makedirs(DST + 'services/vpc')

    for svc, apis in [('ecs', ecs_apis), ('vpc', vpc_apis)]:
        keep = {'client.go'}
        for api in apis:
            fn = find_api_file(svc, api)
            if fn is None:
                sys.exit('!! API not found: %s %s' % (svc, api))
            keep.add(fn)
        for fn in sorted(keep):
            shutil.copy(SRC + f'services/{svc}/{fn}', DST + f'services/{svc}/{fn}')
            os.chmod(DST + f'services/{svc}/{fn}', 0o644)
        print(svc, 'files kept:', len(keep))

    for item in ['sdk', 'go.mod', 'go.sum', 'LICENSE', 'README.md']:
        p = SRC + item
        if os.path.isdir(p):
            shutil.copytree(p, DST + item)
            subprocess.run(['chmod', '-R', 'u+w', DST + item])
        elif os.path.exists(p):
            shutil.copy(p, DST + item)
    print('alibaba fork written to', DST)


def closure():
    """Iteratively add files defining types reported as undefined by the
    compiler (shared struct_*.go helpers), assuming cwd is the autok3s
    module root and the replace directive already points at DST."""
    for _ in range(30):
        r = subprocess.run(['go', 'build', './pkg/providers/alibaba/'],
                           capture_output=True, text=True)
        out = r.stdout + r.stderr
        missing = sorted(set(re.findall(r'undefined: ([A-Za-z0-9_]+)', out)))
        if not missing:
            print('closure clean')
            return
        for typ in missing:
            for svc in ('ecs', 'vpc'):
                d = SRC + f'services/{svc}/'
                for fn in os.listdir(d):
                    if fn == 'client.go' or not fn.endswith('.go'):
                        continue
                    if os.path.exists(DST + f'services/{svc}/{fn}'):
                        continue
                    head = open(d + fn, errors='ignore').read(300000)
                    if re.search(r'\b(type|func) ' + typ + r'\b', head):
                        shutil.copy(d + fn, DST + f'services/{svc}/{fn}')
                        os.chmod(DST + f'services/{svc}/{fn}', 0o644)
                        break
                else:
                    continue
                break


if __name__ == '__main__':
    main()
    closure()
