#!/usr/bin/env python3
"""Drive the accessibility example through the live Linux AT-SPI bus.

Results contain booleans and provenance only. Accessible text and field values
are used for assertions but are never printed or saved in the result record.
"""
import argparse
import ast
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import time

from perf_baseline import provenance

ROOT = Path(__file__).resolve().parent.parent
PREFIX = 'org.a11y.atspi.ownframe.'


def command(args):
    result = subprocess.run(args, cwd=ROOT, capture_output=True, text=True, timeout=10, check=False)
    if result.returncode:
        raise RuntimeError(f'command failed: {args[0]}')
    return result.stdout.strip()


def wait_for(check, label, timeout=30):
    end = time.monotonic() + timeout
    while time.monotonic() < end:
        result = check()
        if result:
            return result
        time.sleep(.1)
    raise RuntimeError(f'timed out: {label}')


def object_path(element):
    return '/org/a11y/atspi/accessible/n' + ('id:' + element).encode().hex()


def variant(output):
    return ast.literal_eval(re.sub(r'\b(?:u?int(?:16|32|64)|byte)\s+', '', output))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--out', type=Path, default=ROOT / 'temp/atspi-actions.json')
    args = parser.parse_args()
    if not shutil.which('gdbus') or not os.environ.get('DISPLAY'):
        parser.error('requires gdbus, a Linux desktop DISPLAY, and an accessibility bus')
    address = ast.literal_eval(command(['gdbus', 'call', '--session', '--dest', 'org.a11y.Bus',
                                       '--object-path', '/org/a11y/bus', '--method', 'org.a11y.Bus.GetAddress']))[0]

    def call(dest, path, method, *values):
        return command(['gdbus', 'call', '--address', address, '--dest', dest,
                        '--object-path', path, '--method', method, '--', *map(str, values)])

    def names():
        output = call('org.freedesktop.DBus', '/org/freedesktop/DBus', 'org.freedesktop.DBus.ListNames')
        return {name for name in ast.literal_eval(output)[0] if name.startswith(PREFIX)}

    before = names()
    args.out.parent.mkdir(parents=True, exist_ok=True)
    report = {**provenance(ROOT), 'status': 'running',
              'go_version': command(['go', 'version']), 'os': os.uname().sysname,
              'kernel': os.uname().release, 'display': os.environ['DISPLAY'],
              'client': 'gdbus', 'host': 'native desktop accessibility example'}
    with args.out.with_suffix('.app.log').open('w') as log:
        process = subprocess.Popen(['go', 'run', './examples/accessibility'], cwd=ROOT,
                                   stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        try:
            def appeared():
                if process.poll() is not None:
                    raise RuntimeError('example exited before registration; inspect the app log')
                return names() - before
            registered = wait_for(appeared, 'new OwnFrame bus registration')
            if len(registered) != 1:
                raise RuntimeError('multiple new OwnFrame applications appeared; rerun without concurrent launches')
            dest = registered.pop()

            def ready():
                try:
                    call(dest, object_path('display-name'), 'org.a11y.atspi.Accessible.GetState')
                    return True
                except RuntimeError:
                    return False
            wait_for(ready, 'published control tree')
            focus = call(dest, object_path('display-name'), 'org.a11y.atspi.Component.GrabFocus')
            if focus != '(true,)':
                raise RuntimeError('live GrabFocus was refused')

            def focused():
                state = variant(call(dest, object_path('display-name'), 'org.a11y.atspi.Accessible.GetState'))[0]
                return bool(state[0] & (1 << 12))
            wait_for(focused, 'focused state publication')
            report['focus_accepted'] = report['focus_state_observed'] = True

            target = object_path('activation-target')
            actions = ast.literal_eval(call(dest, target, 'org.a11y.atspi.Action.GetActions'))[0]
            index = next(i for i, row in enumerate(actions) if row[0] == 'click')
            if call(dest, target, 'org.a11y.atspi.Action.DoAction', index) != '(true,)':
                raise RuntimeError('live activation was refused')

            def activated():
                text = ast.literal_eval(call(dest, target, 'org.a11y.atspi.Accessible.GetName'))[0]
                return '(count: 1)' in text
            wait_for(activated, 'activation counter publication')
            report['activation_accepted'] = report['activation_counter_changed'] = True
            disabled = call(dest, object_path('disabled-field'), 'org.a11y.atspi.Component.GrabFocus')
            if disabled != '(false,)':
                raise RuntimeError('disabled control accepted focus')
            password = ast.literal_eval(call(dest, object_path('secret'), 'org.a11y.atspi.Text.GetText', 0, -1))[0]
            if password:
                raise RuntimeError('password redaction assertion failed')
            report['disabled_focus_refused'] = report['password_redacted'] = True
            report['status'] = 'passed'
        except Exception as error:
            report['status'], report['error'] = 'failed', str(error)
            raise
        finally:
            try:
                os.killpg(process.pid, signal.SIGTERM)
            except ProcessLookupError:
                pass
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
            args.out.write_text(json.dumps(report, indent=2) + '\n')
    print(f'PASS: live focus, activation, disabled control, and password redaction; {args.out}')


if __name__ == '__main__':
    main()
