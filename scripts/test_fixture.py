"""Regression coverage for disposable-fixture startup, without Docker."""
import http.client
import unittest
from unittest.mock import MagicMock, patch
import urllib.error
from fixture import wait_for_health

class ReadinessTests(unittest.TestCase):
    @patch("fixture.time.sleep")
    @patch("fixture.urllib.request.urlopen")
    def test_transient_startup_failures(self, open_url, sleep):
        healthy = MagicMock()
        healthy.__enter__.return_value.status = 200
        open_url.side_effect = [
            ConnectionResetError("Linux container starting"),
            http.client.RemoteDisconnected("Windows container starting"),
            urllib.error.URLError("not listening"),
            TimeoutError("startup timeout"),
            healthy,
        ]
        wait_for_health("http://127.0.0.1:18081")
        self.assertEqual(open_url.call_count, 5)
        self.assertEqual(sleep.call_count, 4)

    @patch("fixture.time.sleep")
    @patch("fixture.time.monotonic", side_effect=[0, 1, 120])
    @patch("fixture.urllib.request.urlopen", side_effect=ConnectionResetError())
    def test_deadline_is_bounded(self, open_url, clock, sleep):
        with self.assertRaisesRegex(RuntimeError, "120 seconds"):
            wait_for_health("http://127.0.0.1:18081")
        self.assertEqual(open_url.call_count, 1)
        sleep.assert_not_called()

    @patch("fixture.time.sleep")
    @patch("fixture.time.monotonic", side_effect=[0, 1, 120])
    @patch("fixture.urllib.request.urlopen")
    def test_non_ready_status_also_obeys_deadline(self, open_url, clock, sleep):
        open_url.return_value.__enter__.return_value.status = 204
        with self.assertRaisesRegex(RuntimeError, "120 seconds"):
            wait_for_health("http://127.0.0.1:18081")
        self.assertEqual(open_url.call_count, 1)
        sleep.assert_not_called()

if __name__ == "__main__":
    unittest.main()
