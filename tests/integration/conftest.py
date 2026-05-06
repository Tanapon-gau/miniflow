import time

import httpx
import pytest

API_URL = "http://localhost:8000"
STARTUP_TIMEOUT = 90  # seconds — covers docker image pulls on first run


@pytest.fixture(scope="session")
def api_client() -> httpx.Client:
    client = httpx.Client(base_url=API_URL, timeout=10, follow_redirects=True)
    deadline = time.time() + STARTUP_TIMEOUT
    while time.time() < deadline:
        try:
            if client.get("/health").status_code == 200:
                return client
        except httpx.ConnectError:
            pass
        time.sleep(2)
    pytest.fail(f"API at {API_URL} did not become healthy within {STARTUP_TIMEOUT}s")
