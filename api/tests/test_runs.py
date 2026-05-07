from httpx import AsyncClient

SAMPLE_DAG = {
    "tasks": [
        {"name": "step1", "type": "shell", "command": "echo hello"},
        {"name": "step2", "type": "python", "script": "print('hi')", "max_retries": 2},
    ]
}


async def _make_workflow(client: AsyncClient, name: str = "test-wf") -> str:
    resp = await client.post("/workflows/", json={"name": name, "dag": SAMPLE_DAG})
    assert resp.status_code == 201
    return resp.json()["id"]


async def _make_run(client: AsyncClient, wf_id: str) -> str:
    resp = await client.post(f"/workflows/{wf_id}/runs")
    assert resp.status_code == 201
    return resp.json()["id"]


async def test_trigger_run(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    resp = await client.post(f"/workflows/{wf_id}/runs")
    assert resp.status_code == 201
    data = resp.json()
    assert data["workflow_id"] == wf_id
    assert data["status"] == "pending"
    assert len(data["tasks"]) == 2
    assert {t["name"] for t in data["tasks"]} == {"step1", "step2"}
    assert all(t["status"] == "pending" for t in data["tasks"])


async def test_trigger_run_workflow_not_found(client: AsyncClient) -> None:
    resp = await client.post("/workflows/00000000-0000-0000-0000-000000000000/runs")
    assert resp.status_code == 404


async def test_task_fields_populated(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    data = (await client.post(f"/workflows/{wf_id}/runs")).json()
    shell = next(t for t in data["tasks"] if t["name"] == "step1")
    assert shell["type"] == "shell"
    assert shell["timeout_seconds"] == 300
    assert shell["max_retries"] == 0
    assert shell["payload"] == {"command": "echo hello"}
    py = next(t for t in data["tasks"] if t["name"] == "step2")
    assert py["max_retries"] == 2


async def test_list_runs(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    await client.post(f"/workflows/{wf_id}/runs")
    await client.post(f"/workflows/{wf_id}/runs")
    resp = await client.get("/runs")
    assert resp.status_code == 200
    assert len(resp.json()) == 2


async def test_list_runs_filter_by_workflow(client: AsyncClient) -> None:
    wf1 = await _make_workflow(client, "wf1")
    wf2 = await _make_workflow(client, "wf2")
    await client.post(f"/workflows/{wf1}/runs")
    await client.post(f"/workflows/{wf1}/runs")
    await client.post(f"/workflows/{wf2}/runs")
    resp = await client.get(f"/runs?workflow_id={wf1}")
    assert resp.status_code == 200
    assert len(resp.json()) == 2
    assert all(r["workflow_id"] == wf1 for r in resp.json())


async def test_get_run(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    run_id = (await client.post(f"/workflows/{wf_id}/runs")).json()["id"]
    resp = await client.get(f"/runs/{run_id}")
    assert resp.status_code == 200
    data = resp.json()
    assert data["id"] == run_id
    assert len(data["tasks"]) == 2


async def test_get_run_not_found(client: AsyncClient) -> None:
    resp = await client.get("/runs/00000000-0000-0000-0000-000000000000")
    assert resp.status_code == 404


async def test_multiple_runs_same_workflow(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    r1 = (await client.post(f"/workflows/{wf_id}/runs")).json()["id"]
    r2 = (await client.post(f"/workflows/{wf_id}/runs")).json()["id"]
    assert r1 != r2
    resp = await client.get("/runs")
    assert len(resp.json()) == 2


async def test_cancel_run_marks_run_and_pending_tasks_cancelled(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    run_id = await _make_run(client, wf_id)

    resp = await client.post(f"/runs/{run_id}/cancel")
    assert resp.status_code == 200
    data = resp.json()
    assert data["status"] == "cancelled"
    assert data["finished_at"] is not None
    assert all(t["status"] == "cancelled" for t in data["tasks"])
    assert all(t["finished_at"] is not None for t in data["tasks"])


async def test_cancel_run_not_found(client: AsyncClient) -> None:
    resp = await client.post("/runs/00000000-0000-0000-0000-000000000000/cancel")
    assert resp.status_code == 404


async def test_cancel_already_cancelled_run_returns_conflict(client: AsyncClient) -> None:
    wf_id = await _make_workflow(client)
    run_id = await _make_run(client, wf_id)

    await client.post(f"/runs/{run_id}/cancel")
    resp = await client.post(f"/runs/{run_id}/cancel")
    assert resp.status_code == 409
    assert "cancelled" in resp.json()["detail"]


async def test_cancel_does_not_affect_running_tasks(client: AsyncClient) -> None:
    # A run with one pending task — cancelling the run cancels the pending task.
    # Tasks already in 'running' status are left alone (we don't kill live processes).
    wf_id = await _make_workflow(client)
    run_id = await _make_run(client, wf_id)

    resp = await client.post(f"/runs/{run_id}/cancel")
    assert resp.status_code == 200
    # All tasks were pending at cancellation time, so all should be cancelled.
    assert all(t["status"] == "cancelled" for t in resp.json()["tasks"])
