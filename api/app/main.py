from fastapi import FastAPI

from .routers.workflows import router as workflows_router

app = FastAPI(title="MiniFlow API")
app.include_router(workflows_router)


@app.get("/health")
async def health() -> dict[str, str]:
    return {"status": "ok"}
