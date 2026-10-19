#!/usr/bin/env python3
"""Build a pruned fork of google.golang.org/api for autok3s.

Usage:
    hack/sdk-prune/build-google-api-fork.py <dst-dir> [<src-module-dir>]

Keeps the shared packages (googleapi, internal, option, gensupport,
transport) plus a pruned compute/v1: only the resource services and
verbs autok3s calls survive (the 210k-line compute-gen.go shrinks by
~145k lines); model types are kept wholesale.

Keep the resource/verb matrix below in sync with pkg/providers/google
and pkg/server/store/credential and re-run when bumping the module.
"""

import os
import re
import shutil
import subprocess
import sys

DST = sys.argv[1].rstrip('/') + '/' if len(sys.argv) > 1 else './thirdparty/google-api/'

KEEP_RESOURCES = {'DiskTypes', 'Disks', 'Firewalls', 'GlobalOperations', 'Images',
                  'Instances', 'MachineTypes', 'Networks', 'Projects', 'Regions',
                  'ZoneOperations', 'Zones'}
KEEP_VERBS = {('DiskTypes', 'List'), ('Disks', 'Delete'), ('Disks', 'Get'),
              ('Firewalls', 'Get'), ('Firewalls', 'Insert'), ('Firewalls', 'Update'),
              ('GlobalOperations', 'Get'), ('Images', 'List'), ('Instances', 'Delete'),
              ('Instances', 'Get'), ('Instances', 'Insert'), ('Instances', 'List'),
              ('Instances', 'SetMetadata'), ('MachineTypes', 'List'), ('Networks', 'List'),
              ('Projects', 'Get'), ('Regions', 'List'), ('ZoneOperations', 'Get'),
              ('Zones', 'List')}

if len(sys.argv) > 2:
    SRC = sys.argv[2].rstrip('/') + '/'
else:
    MC = subprocess.run(['go', 'env', 'GOMODCACHE'], capture_output=True, text=True).stdout.strip()
    SRC = MC + '/google.golang.org/api@v0.153.0/'


def prune_compute_gen():
    g = SRC + 'compute/v1/compute-gen.go'
    lines = open(g, errors='ignore').read().split('\n')
    starts = [i for i, l in enumerate(lines) if re.match(r'^(func|type|var|const) ', l)]
    ranges = [(s, (starts[k + 1] - 1 if k + 1 < len(starts) else len(lines) - 1))
              for k, s in enumerate(starts)]
    drop = set()
    for s, e in ranges:
        h = lines[s]
        bad = False
        m = re.search(r'\b([A-Z][A-Za-z0-9]*?)Call\b', h)
        if m:
            call = m.group(1)
            for r in sorted(KEEP_RESOURCES, key=len, reverse=True):
                if call.startswith(r):
                    verb = call[len(r):]
                    if verb and (r, verb) not in KEEP_VERBS:
                        bad = True
                    break
            else:
                bad = True
        if not bad:
            m2 = re.match(r'^func New([A-Z][A-Za-z0-9]*)Service\(s \*Service\)', h)
            m4 = re.match(r'^func \(r \*([A-Z][A-Za-z0-9]*)Service\) ([A-Za-z0-9]+\()', h)
            if h.startswith('type ') and s + 1 < len(lines) and lines[s + 1] == '\ts *Service':
                m5 = re.match(r'^type ([A-Z][A-Za-z0-9]*)Service struct \{', h)
            else:
                m5 = None
            res = verb = None
            if m2:
                res = m2.group(1)
            elif m4:
                res, verb = m4.group(1), m4.group(2).rstrip('(')
            elif m5:
                res = m5.group(1)
            if res:
                if res not in KEEP_RESOURCES:
                    bad = True
                elif m4 and (res, verb) not in KEEP_VERBS:
                    bad = True
        if bad:
            drop.update(range(s, e + 1))
    svc_field = re.compile(r'^\t([A-Z][A-Za-z0-9]*) \*\1Service$')
    new_line = re.compile(r'^\ts\.([A-Z][A-Za-z0-9]*) = New\1Service\(s\)$')
    out = []
    for i, l in enumerate(lines):
        if i in drop:
            continue
        m = svc_field.match(l) or new_line.match(l)
        if m and m.group(1) not in KEEP_RESOURCES:
            continue
        out.append(l)
    return '\n'.join(out)


def main():
    shutil.rmtree(DST, ignore_errors=True)
    os.makedirs(DST + 'compute/v1')
    open(DST + 'compute/v1/compute-gen.go', 'w').write(prune_compute_gen())
    for item in ['go.mod', 'go.sum', 'LICENSE', 'README.md']:
        p = SRC + item
        if os.path.exists(p):
            shutil.copy(p, DST + item)
    for d in ['googleapi', 'internal', 'option', 'gensupport', 'transport']:
        p = SRC + d
        if os.path.isdir(p):
            shutil.copytree(p, DST + d)
            subprocess.run(['chmod', '-R', 'u+w', DST + d])
    subprocess.run('find %s -name "*_test.go" -delete; find %s -name "*.json" -delete' % (DST, DST), shell=True)
    print('google-api fork written to', DST)


if __name__ == '__main__':
    main()
