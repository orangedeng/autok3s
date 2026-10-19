#!/usr/bin/env python3
"""Prune the forked aws-sdk-go-v2 ec2 module to the operations autok3s uses.

Usage:
    hack/sdk-prune/prune-aws-ec2.py <path-to-ec2-module>

Steps (run against a pristine upstream module before applying this script):
  1. delete api_op_*.go files whose operation is not listed in USED_OPS below
  2. drop test fixtures (*_test.go, request_snapshot/, snapshot/, CHANGELOG.md)
  3. strip per-operation blocks from serializers.go / deserializers.go /
     validators.go (this script)

The script splits files at top-level declarations (gofmt guarantees
column-0 separation; brace counting is deliberately avoided because
generated code carries braces inside string literals) and removes every
block belonging to an unused operation.

When bumping aws-sdk-go-v2, re-copy the pristine module over
thirdparty/ec2, re-apply step 1 and 2, then run this script. If the
smithy codegen shapes change, adjust the regexes below.
"""

import os
import re
import sys

# Operations called by pkg/providers/aws (plus the instance-running waiter,
# which lives in api_op_DescribeInstances.go).
USED_OPS = {
    'AuthorizeSecurityGroupIngress',
    'CreateSecurityGroup',
    'CreateTags',
    'DeleteTags',
    'DescribeAccountAttributes',
    'DescribeImages',
    'DescribeInstances',
    'DescribeKeyPairs',
    'DescribeSecurityGroups',
    'DescribeSubnets',
    'ImportKeyPair',
    'RunInstances',
    'TerminateInstances',
}

OPRE = re.compile(r'awsEc2query_(?:serialize|deserialize)Op(?:Document)?([A-Za-z0-9]+)')
VALRES = [
    re.compile(r'^func validateOp([A-Za-z0-9]+)'),
    re.compile(r'^func \((?:\w+ )?\*?validateOp([A-Za-z0-9]+)\)'),
    re.compile(r'^type validateOp([A-Za-z0-9]+) struct'),
    re.compile(r'^func addOp([A-Za-z0-9]+)ValidationMiddleware'),
]
HEAD = re.compile(r'^(func|type|var|const) ')


def norm(name):
    for suf in ('Request', 'Input', 'Output'):
        if name.endswith(suf):
            return name[:-len(suf)]
    if name.startswith('Error'):
        return name[len('Error'):]
    return name


def prune(path):
    lines = open(path).read().split('\n')
    starts = [i for i, l in enumerate(lines) if HEAD.match(l)]
    ranges = [(s, (starts[k + 1] - 1 if k + 1 < len(starts) else len(lines) - 1))
              for k, s in enumerate(starts)]
    drop = set()
    for s, e in ranges:
        header = lines[s]
        bad = False
        for vr in VALRES:
            m = vr.match(header)
            if m and norm(m.group(1)) not in USED_OPS:
                bad = True
                break
        if not bad:
            m = OPRE.search(header)
            if m and norm(m.group(1)) not in USED_OPS:
                bad = True
        if bad:
            drop.update(range(s, e + 1))
    open(path, 'w').write('\n'.join(l for i, l in enumerate(lines) if i not in drop))
    print('%s: dropped %d lines' % (path, len(drop)))


def main():
    base = sys.argv[1].rstrip('/') + '/'
    for f in ('serializers.go', 'deserializers.go', 'validators.go'):
        prune(os.path.join(base, f))


if __name__ == '__main__':
    main()
