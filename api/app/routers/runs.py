import uuid

from fastapi import APIRouter, Depends, HTTPException, Query, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from ..deps import get_session
from ..models import Run, Task, Workflow
from ..schemas import RunDetail, RunRead

router = APIRouter(tags=["runs"])

_TASK_META_KEYS = {"name", "type", "timeout_seconds", "max_retries"}


@router.post(
    "/workflows/{workflow_id}/runs",
    response_model=RunDetail,
    status_code=status.HTTP_201_CREATED,
)
async def trigger_run(
    workflow_id: uuid.UUID,
    session: AsyncSession = Depends(get_session),
) -> Run:
    workflow = await session.get(Workflow, workflow_id)
    if workflow is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"workflow {workflow_id} not found",
        )

    run = Run(workflow_id=workflow_id)
    session.add(run)
    await session.flush()

    for task_def in workflow.dag.get("tasks", []):
        task = Task(
            run_id=run.id,
            name=task_def["name"],
            type=task_def["type"],
            payload={k: v for k, v in task_def.items() if k not in _TASK_META_KEYS},
            timeout_seconds=task_def.get("timeout_seconds", 300),
            max_retries=task_def.get("max_retries", 0),
        )
        session.add(task)

    await session.commit()

    result = await session.execute(
        select(Run).where(Run.id == run.id).options(selectinload(Run.tasks))
    )
    return result.scalar_one()


@router.get("/runs", response_model=list[RunRead])
async def list_runs(
    workflow_id: uuid.UUID | None = Query(default=None),
    session: AsyncSession = Depends(get_session),
) -> list[Run]:
    stmt = select(Run).order_by(Run.triggered_at.desc())
    if workflow_id is not None:
        stmt = stmt.where(Run.workflow_id == workflow_id)
    result = await session.execute(stmt)
    return list(result.scalars().all())


@router.get("/runs/{run_id}", response_model=RunDetail)
async def get_run(
    run_id: uuid.UUID,
    session: AsyncSession = Depends(get_session),
) -> Run:
    result = await session.execute(
        select(Run).where(Run.id == run_id).options(selectinload(Run.tasks))
    )
    run = result.scalar_one_or_none()
    if run is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"run {run_id} not found",
        )
    return run
