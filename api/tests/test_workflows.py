from httpx import AsyncClient

SAMPLE_DAG = {"tasks": [{"name": "echo", "type": "shell", "command": "echo hello"}]}


async def test_create_workflow(client: AsyncClient) -> None:
    resp = await client.post("/workflows/", json={"name": "my-wf", "dag": SAMPLE_DAG})
    assert resp.status_code == 201
    data = resp.json()
    assert data["name"] == "my-wf"
    assert data["dag"] == SAMPLE_DAG
    assert "id" in data


async def test_list_workflows(client: AsyncClient) -> None:
    await client.post("/workflows/", json={"name": "wf1", "dag": SAMPLE_DAG})
    await client.post("/workflows/", json={"name": "wf2", "dag": SAMPLE_DAG})
    resp = await client.get("/workflows/")
    assert resp.status_code == 200
    assert len(resp.json()) == 2


async def test_get_workflow(client: AsyncClient) -> None:
    create = await client.post("/workflows/", json={"name": "wf", "dag": SAMPLE_DAG})
    wf_id = create.json()["id"]
    resp = await client.get(f"/workflows/{wf_id}")
    assert resp.status_code == 200
    assert resp.json()["id"] == wf_id


async def test_get_workflow_not_found(client: AsyncClient) -> None:
    resp = await client.get("/workflows/00000000-0000-0000-0000-000000000000")
    assert resp.status_code == 404


async def test_update_workflow(client: AsyncClient) -> None:
    create = await client.post("/workflows/", json={"name": "old", "dag": SAMPLE_DAG})
    wf_id = create.json()["id"]
    resp = await client.put(f"/workflows/{wf_id}", json={"name": "new"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["name"] == "new"
    assert body["dag"] == SAMPLE_DAG  # unchanged by partial update


async def test_update_workflow_not_found(client: AsyncClient) -> None:
    resp = await client.put(
        "/workflows/00000000-0000-0000-0000-000000000000", json={"name": "x"}
    )
    assert resp.status_code == 404


async def test_delete_workflow(client: AsyncClient) -> None:
    create = await client.post("/workflows/", json={"name": "wf", "dag": SAMPLE_DAG})
    wf_id = create.json()["id"]
    assert (await client.delete(f"/workflows/{wf_id}")).status_code == 204
    assert (await client.get(f"/workflows/{wf_id}")).status_code == 404


async def test_delete_workflow_not_found(client: AsyncClient) -> None:
    resp = await client.delete("/workflows/00000000-0000-0000-0000-000000000000")
    assert resp.status_code == 404


async def test_create_with_empty_name(client: AsyncClient) -> None:
    resp = await client.post("/workflows/", json={"name": "  ", "dag": SAMPLE_DAG})
    assert resp.status_code == 422


async def test_create_dag_without_tasks_key(client: AsyncClient) -> None:
    resp = await client.post("/workflows/", json={"name": "wf", "dag": {}})
    assert resp.status_code == 422
