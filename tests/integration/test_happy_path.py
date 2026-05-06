import time

import httpx
import pytest

POLL_INTERVAL = 2   # seconds between status checks
COMPLETION_TIMEOUT = 60  # seconds to wait for a run to reach terminal status


def wait_for_completion(client: httpx.Client, run_id: str) -> dict:
    deadline = time.time() + COMPLETION_TIMEOUT
    while time.time() < deadline:
        run = client.get(f"/runs/{run_id}").json()
        if run["status"] in ("success", "failed"):
            return run
        time.sleep(POLL_INTERVAL)
    pytest.fail(f"run {run_id} did not complete within {COMPLETION_TIMEOUT}s")


def create_workflow(client: httpx.Client, dag: dict) -> str:
    response = client.post("/workflows", json={"name": "e2e-test", "dag": dag})
    assert response.status_code == 201, response.text
    return response.json()["id"]


def trigger_run(client: httpx.Client, workflow_id: str) -> str:
    response = client.post(f"/workflows/{workflow_id}/runs")
    assert response.status_code == 201, response.text
    return response.json()["id"]


class TestShellTask:
    def test_single_task_succeeds(self, api_client: httpx.Client) -> None:
        dag = {
            "tasks": [
                {
                    "name": "say_hello",
                    "type": "shell",
                    "command": "echo hello",
                    "timeout_seconds": 10,
                }
            ]
        }
        run_id = trigger_run(api_client, create_workflow(api_client, dag))
        run = wait_for_completion(api_client, run_id)

        assert run["status"] == "success", f"run ended with status {run['status']!r}"
        assert len(run["tasks"]) == 1
        assert run["tasks"][0]["status"] == "success"

    def test_two_task_dag_respects_dependency_order(self, api_client: httpx.Client) -> None:
        # task_b depends on task_a — scheduler must not dispatch b until a succeeds
        dag = {
            "tasks": [
                {
                    "name": "task_a",
                    "type": "shell",
                    "command": "echo step_a",
                    "timeout_seconds": 10,
                },
                {
                    "name": "task_b",
                    "type": "shell",
                    "command": "echo step_b",
                    "depends_on": ["task_a"],
                    "timeout_seconds": 10,
                },
            ]
        }
        run_id = trigger_run(api_client, create_workflow(api_client, dag))
        run = wait_for_completion(api_client, run_id)

        assert run["status"] == "success", f"run ended with status {run['status']!r}"
        by_name = {t["name"]: t for t in run["tasks"]}
        assert by_name["task_a"]["status"] == "success"
        assert by_name["task_b"]["status"] == "success"
        # task_b must have started after task_a finished (dependency ordering)
        from datetime import datetime
        b_started = datetime.fromisoformat(by_name["task_b"]["started_at"])
        a_finished = datetime.fromisoformat(by_name["task_a"]["finished_at"])
        assert b_started >= a_finished, (
            f"task_b started at {b_started} before task_a finished at {a_finished}"
        )

    def test_failing_task_marks_run_failed(self, api_client: httpx.Client) -> None:
        dag = {
            "tasks": [
                {
                    "name": "will_fail",
                    "type": "shell",
                    "command": "exit 1",
                    "timeout_seconds": 10,
                }
            ]
        }
        run_id = trigger_run(api_client, create_workflow(api_client, dag))
        run = wait_for_completion(api_client, run_id)

        assert run["status"] == "failed"
        assert run["tasks"][0]["status"] == "failed"
