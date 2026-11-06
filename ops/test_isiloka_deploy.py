import contextlib
import io
from pathlib import Path
import runpy
import sys
import types
import unittest
from unittest.mock import MagicMock, patch


SCRIPT = Path(__file__).with_name('isiloka-activate.py')


class DeploymentPreflightTests(unittest.TestCase):
    def check(self, *, machine='x86_64', inactive=None, env_exists=True,
              health_error=None, args=None):
        calls = []

        def service_status(command, **kwargs):
            calls.append(command)
            return types.SimpleNamespace(returncode=int(command[-1] == inactive))

        opener = MagicMock()
        opener.open.return_value.__enter__.return_value.read.return_value = b'ok'
        opener.open.side_effect = health_error
        with contextlib.ExitStack() as stack:
            stack.enter_context(patch.dict(sys.modules, {'fcntl': MagicMock()}))
            stack.enter_context(patch.object(sys, 'argv', [str(SCRIPT)] + (args or ['--check'])))
            stack.enter_context(patch('platform.machine', return_value=machine))
            stack.enter_context(patch('subprocess.run', side_effect=service_status))
            stack.enter_context(patch('pathlib.Path.is_file', return_value=env_exists))
            stack.enter_context(patch('urllib.request.build_opener', return_value=opener))
            file_open = stack.enter_context(patch('builtins.open'))
            output = stack.enter_context(contextlib.redirect_stdout(io.StringIO()))
            try:
                runpy.run_path(str(SCRIPT), run_name='__main__')
            except (SystemExit, OSError) as exc:
                result = exc
            else:
                self.fail('Preflight must exit without activating a release')
            file_open.assert_not_called()
        self.assertTrue(all(c[:2] == ['systemctl', 'is-active'] for c in calls))
        return result, calls, opener, output.getvalue()

    def test_restored_server_passes_without_mutations(self):
        result, calls, opener, output = self.check()
        self.assertEqual(result.code, 0)
        self.assertEqual([c[-1] for c in calls], ['isiloka-web', 'isiloka-be'])
        self.assertEqual([c.args[0] for c in opener.open.call_args_list], [
            'http://127.0.0.1:8102/health', 'http://127.0.0.1:33128/',
        ])
        self.assertIn('prerequisites OK', output)

    def test_missing_backend_stops_before_health_checks(self):
        result, _, opener, _ = self.check(inactive='isiloka-be')
        self.assertIn('Restore the database/backend', str(result))
        opener.open.assert_not_called()

    def test_wrong_architecture_stops_before_service_checks(self):
        result, calls, _, _ = self.check(machine='aarch64')
        self.assertIn('x86_64', str(result))
        self.assertEqual(calls, [])

    def test_missing_environment_stops(self):
        result, _, opener, _ = self.check(env_exists=False)
        self.assertIn('Missing server environment', str(result))
        opener.open.assert_not_called()

    def test_failed_health_check_stops(self):
        result, _, _, _ = self.check(health_error=OSError('unavailable'))
        self.assertIsInstance(result, OSError)

    def test_unknown_arguments_stop_before_service_checks(self):
        result, calls, _, _ = self.check(args=['--wrong'])
        self.assertIn('Usage:', str(result))
        self.assertEqual(calls, [])


if __name__ == '__main__':
    unittest.main()
